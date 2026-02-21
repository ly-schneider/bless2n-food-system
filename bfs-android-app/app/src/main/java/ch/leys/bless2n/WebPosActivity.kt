package ch.leys.bless2n

import android.Manifest
import android.annotation.SuppressLint
import android.content.Intent
import android.content.BroadcastReceiver
import android.content.IntentFilter
import android.content.pm.PackageManager
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.webkit.JavascriptInterface
import android.webkit.WebChromeClient
import android.webkit.WebResourceRequest
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.activity.ComponentActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import androidx.webkit.WebSettingsCompat
import androidx.webkit.WebViewFeature
import android.webkit.CookieManager
import android.bluetooth.BluetoothManager
import android.bluetooth.BluetoothDevice
import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.Color
import android.graphics.Paint
import com.google.zxing.BarcodeFormat
import com.google.zxing.common.BitMatrix
import com.google.zxing.qrcode.QRCodeWriter
import com.sumup.merchant.reader.api.SumUpAPI
import com.sumup.merchant.reader.api.SumUpLogin
import com.sumup.merchant.reader.api.SumUpPayment
import java.math.BigDecimal
import org.json.JSONObject
import android.widget.Toast
import android.util.Log
import java.util.UUID
import org.json.JSONArray

import android.bluetooth.BluetoothAdapter
import android.content.Context

class WebPosActivity : ComponentActivity() {

    companion object {
        private const val REQ_LOGIN = 1001
        private const val REQ_CHECKOUT = 1002
    }

    private lateinit var webView: WebView
    private var pendingPayment: PendingPayment? = null
    // Store last successful SumUp transaction details by correlationId/reference
    private val txByCorrelation: MutableMap<String, JSONObject> = mutableMapOf()
    // Last successful transaction (fallback when correlationId is not provided)
    private var lastTx: JSONObject? = null

    // Simple prefs for storing selected printer
    private val prefs by lazy { getSharedPreferences("pos_prefs", Context.MODE_PRIVATE) }
    private fun saveSelectedPrinter(mac: String?) {
        prefs.edit().putString("printer_mac", mac).apply()
    }
    private fun loadSelectedPrinter(): String? = prefs.getString("printer_mac", null)

    // Bluetooth discovery receiver lifecycle
    private var btReceiver: BroadcastReceiver? = null
    private var discoveryActive: Boolean = false
    private fun registerBtReceiver() {
        if (btReceiver != null) return
        btReceiver = object : BroadcastReceiver() {
            override fun onReceive(context: Context?, intent: Intent?) {
                when (intent?.action) {
                    BluetoothDevice.ACTION_FOUND -> {
                        try {
                            val dev: BluetoothDevice? = intent.getParcelableExtra(BluetoothDevice.EXTRA_DEVICE)
                            if (dev != null) {
                                val name = try { dev.name ?: "Bluetooth Device" } catch (_: SecurityException) { "Bluetooth Device" } catch (_: Throwable) { "Bluetooth Device" }
                                val addr = try { dev.address ?: "" } catch (_: SecurityException) { "" } catch (_: Throwable) { "" }
                                val bonded = try { dev.bondState == BluetoothDevice.BOND_BONDED } catch (_: Throwable) { false }
                                val proto = if (name.contains("T02", true) || name.contains("Phomemo", true)) "phomemo" else "escpos"
                                if (addr.isNotBlank()) {
                                    sendEvent(
                                        "bfs:printer:found",
                                        JSONObject().put("name", name).put("address", addr).put("bonded", bonded).put("protocol", proto)
                                    )
                                }
                            }
                        } catch (_: Throwable) {}
                    }
                    BluetoothDevice.ACTION_BOND_STATE_CHANGED -> {
                        val dev: BluetoothDevice? = intent.getParcelableExtra(BluetoothDevice.EXTRA_DEVICE)
                        if (dev != null) {
                            val addr = try { dev.address ?: "" } catch (_: Throwable) { "" }
                            val state = when (dev.bondState) {
                                BluetoothDevice.BOND_BONDING -> "bonding"
                                BluetoothDevice.BOND_BONDED -> "bonded"
                                else -> "none"
                            }
                            if (addr.isNotBlank()) {
                                sendEvent(
                                    "bfs:printer:bond:state",
                                    JSONObject().put("address", addr).put("state", state)
                                )
                            }
                        }
                    }
                    BluetoothAdapter.ACTION_DISCOVERY_FINISHED -> {
                        discoveryActive = false
                        sendEvent("bfs:printer:discovery:finished", JSONObject())
                    }
                }
            }
        }
        val f = IntentFilter().apply {
            addAction(BluetoothDevice.ACTION_FOUND)
            addAction(BluetoothDevice.ACTION_BOND_STATE_CHANGED)
            addAction(BluetoothAdapter.ACTION_DISCOVERY_FINISHED)
        }
        try { registerReceiver(btReceiver, f) } catch (_: Throwable) {}
    }

    private fun unregisterBtReceiver() {
        val r = btReceiver ?: return
        btReceiver = null
        try { unregisterReceiver(r) } catch (_: Throwable) {}
    }

    data class PendingPayment(
        val amountCents: Long,
        val currency: String,
        val reference: String
    )

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        webView = WebView(this)
        configureWebView(webView)
        setContentView(webView)

        // Bridge for PosBridge.* calls from /pos page
        webView.addJavascriptInterface(PosBridge(), "PosBridge")

        webView.loadUrl(BuildConfig.POS_URL)

        AppUpdateChecker(this).checkForUpdate()
    }

    override fun onDestroy() {
        super.onDestroy()
        try {
            val mgr = getSystemService(BLUETOOTH_SERVICE) as BluetoothManager
            val adapter = mgr.adapter
            if (adapter != null && adapter.isDiscovering) {
                try { adapter.cancelDiscovery() } catch (_: Throwable) {}
            }
        } catch (_: Throwable) {}
        unregisterBtReceiver()
    }

    @SuppressLint("SetJavaScriptEnabled")
    private fun configureWebView(wv: WebView) {
        val s: WebSettings = wv.settings
        s.javaScriptEnabled = true
        s.domStorageEnabled = true
        s.databaseEnabled = false
        s.allowFileAccess = false
        s.allowContentAccess = false
        s.javaScriptCanOpenWindowsAutomatically = false
        s.mediaPlaybackRequiresUserGesture = true
        s.cacheMode = WebSettings.LOAD_DEFAULT

        if (WebViewFeature.isFeatureSupported(WebViewFeature.SAFE_BROWSING_ENABLE)) {
            WebSettingsCompat.setSafeBrowsingEnabled(s, true)
        }

        // Allow mixed content only in dev builds to support http://10.0.2.2
        if (BuildConfig.DEV_BUILD) {
            s.mixedContentMode = WebSettings.MIXED_CONTENT_COMPATIBILITY_MODE
        } else {
            s.mixedContentMode = WebSettings.MIXED_CONTENT_NEVER_ALLOW
        }

        if (WebViewFeature.isFeatureSupported(WebViewFeature.FORCE_DARK)) {
            WebSettingsCompat.setForceDark(s, WebSettingsCompat.FORCE_DARK_AUTO)
        }

        val allowedHost = Uri.parse(BuildConfig.POS_URL).host

        // Allow cookies for cross-domain auth flows; third-party cookies only in dev
        val cm = CookieManager.getInstance()
        cm.setAcceptCookie(true)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
            cm.setAcceptThirdPartyCookies(wv, BuildConfig.DEV_BUILD)
        }

        wv.webViewClient = object : WebViewClient() {
            override fun shouldOverrideUrlLoading(
                view: WebView,
                request: WebResourceRequest
            ): Boolean {
                val url = request.url
                val host = url.host
                // Allow in-app only the configured host; open others externally
                return if (host != null && host == allowedHost) {
                    false
                } else {
                    startActivity(Intent(Intent.ACTION_VIEW, url))
                    true
                }
            }
        }
        wv.webChromeClient = WebChromeClient()
    }

    inner class PosBridge {
        // Called from JS: PosBridge.payWithCard({ amountCents, currency, reference })
        @JavascriptInterface
        fun payWithCard(json: String) {
            var ref: String = ""
            try {
                val obj = try { JSONObject(json) } catch (_: Throwable) { null }
                if (obj == null) {
                    sendEvent(
                        "bfs:sumup:result",
                        JSONObject().put("success", false).put("error", "invalid_request").put("correlationId", "")
                    )
                    return
                }
                val amountCents = obj.optLong("amountCents", -1L)
                val currency = obj.optString("currency", "CHF")
                ref = obj.optString("reference", "pos_${System.currentTimeMillis()}")
                if (amountCents <= 0) {
                    sendEvent(
                        "bfs:sumup:result",
                        JSONObject().put("success", false).put("error", "invalid_amount").put("correlationId", ref)
                    )
                    return
                }
                val pending = PendingPayment(amountCents, currency, ref)
                runOnUiThread { startSumUpCheckout(pending) }
            } catch (e: Exception) {
                sendEvent(
                    "bfs:sumup:result",
                    JSONObject().put("success", false).put("error", e.message ?: "invalid_request").put("correlationId", ref)
                )
            }
        }

        // Called from JS: PosBridge.print(jsonStringOrObject)
        @JavascriptInterface
        fun print(payload: String) {
            val correlationId = try {
                val obj = JSONObject(payload)
                obj.optString("correlationId", "print_${System.currentTimeMillis()}")
            } catch (_: Exception) {
                "print_${System.currentTimeMillis()}"
            }

            // For MVP: use system print for HTML, ESC/POS for plain receipt
            try {
                val obj = try { JSONObject(payload) } catch (_: Exception) { JSONObject() }

                // Skip printing silently if explicitly requested (e.g., Jeton mode)
                val skipPrint = obj.optBoolean("skipPrint", false)
                if (skipPrint) {
                    sendEvent(
                        "bfs:print:result",
                        JSONObject().put("success", true).put("skipped", true).put("correlationId", correlationId)
                    )
                    return
                }

                val mode = obj.optString("mode", "escpos")
                if (mode == "system") {
                    val html = obj.optString("content", payload)
                    printWithSystem(html) { success, error ->
                        sendEvent(
                            "bfs:print:result",
                            JSONObject().put("success", success).put("error", error ?: JSONObject.NULL).put("correlationId", correlationId)
                        )
                    }
                } else {
                    // ESC/POS (and Phomemo fallback) path
                    // Try to augment receipt content with SumUp tx details if available
                    val rawContent = obj.optString("content", payload)
                    // Optional explicit printer selection in payload
                    val explicitPrinterAddr = obj.optString("printerAddress", obj.optString("printer", "")).ifBlank { null }
                    val augmentedContent = try {
                        val rc = JSONObject(rawContent)
                        val method = rc.optString("method", "").lowercase()
                        if (method == "card" && !rc.has("cardLast4")) {
                            val tx = txByCorrelation[correlationId] ?: lastTx
                            if (tx != null) {
                                // Copy over known non-sensitive card + tx fields
                                tx.optString("cardLast4").takeIf { it.isNotEmpty() }?.let { rc.put("cardLast4", it) }
                                tx.optString("cardType").takeIf { it.isNotEmpty() }?.let { rc.put("cardType", it) }
                                tx.optString("entryMode").takeIf { it.isNotEmpty() }?.let { rc.put("entryMode", it) }
                                tx.optString("txId").takeIf { it.isNotEmpty() }?.let { rc.put("txId", it) }
                                tx.optString("status").takeIf { it.isNotEmpty() }?.let { rc.put("txStatus", it) }
                                // Prefer currency from tx; keep existing if already set
                                if (!rc.has("currency")) tx.optString("currency").takeIf { it.isNotEmpty() }?.let { rc.put("currency", it) }
                                if (!rc.has("totalCents")) tx.optLong("amountCents", -1L).takeIf { it >= 0 }?.let { rc.put("totalCents", it) }
                            }
                        }
                        rc.toString()
                    } catch (_: Throwable) {
                        rawContent
                    }
                    val ok = printWithEscPos(augmentedContent, explicitPrinterAddr)
                    sendEvent(
                        "bfs:print:result",
                        JSONObject().put("success", ok).put("correlationId", correlationId)
                    )
                }
            } catch (e: Exception) {
                sendEvent(
                    "bfs:print:result",
                    JSONObject().put("success", false).put("error", e.message ?: "print failed").put("correlationId", correlationId)
                )
            }
        }

        // Return a JSON array string of paired Bluetooth printers the app can see
        // [{ name, address, protocol }]
        @JavascriptInterface
        fun listPrinters(): String {
            if (!ensureBtPermissions()) return "[]"
            val arr = JSONArray()
            try {
                // Prefer DantSu helper to enumerate known printer connections
                val dsuList = try {
                    com.dantsu.escposprinter.connection.bluetooth.BluetoothPrintersConnections().list
                } catch (_: Throwable) { null }
                if (dsuList != null && dsuList.isNotEmpty()) {
                    for (conn in dsuList) {
                        try {
                            val dev = try { conn.device } catch (_: Throwable) { null }
                            val name = try { dev?.name ?: "Bluetooth Printer" } catch (_: Throwable) { "Bluetooth Printer" }
                            val addr = try { dev?.address ?: "" } catch (_: Throwable) { "" }
                            val proto = if (name.contains("T02", true) || name.contains("Phomemo", true)) "phomemo" else "escpos"
                            if (addr.isNotBlank()) {
                                arr.put(
                                    JSONObject().put("name", name).put("address", addr).put("protocol", proto)
                                )
                            }
                        } catch (_: Throwable) { /* ignore entry */ }
                    }
                } else {
                    // Fallback: list bonded devices from adapter
                    val mgr = getSystemService(BLUETOOTH_SERVICE) as BluetoothManager
                    val adapter = mgr.adapter
                    val bonded = try { adapter?.bondedDevices } catch (_: SecurityException) { null }
                    bonded?.forEach { dev ->
                        try {
                            val name = try { dev.name } catch (_: Throwable) { "Bluetooth Device" }
                            val addr = try { dev.address } catch (_: Throwable) { "" }
                            if (addr.isNotBlank()) {
                                val proto = if (name.contains("T02", true) || name.contains("Phomemo", true)) "phomemo" else "escpos"
                                arr.put(JSONObject().put("name", name).put("address", addr).put("protocol", proto))
                            }
                        } catch (_: Throwable) { }
                    }
                }
            } catch (_: Throwable) { }
            return arr.toString()
        }

        // Persist selected printer MAC address (null or empty clears selection)
        @JavascriptInterface
        fun selectPrinter(address: String?) {
            val mac = address?.trim().orEmpty()
            if (mac.isBlank()) saveSelectedPrinter(null) else saveSelectedPrinter(mac)
        }

        // Return current selected printer MAC (or empty)
        @JavascriptInterface
        fun getSelectedPrinter(): String {
            return loadSelectedPrinter() ?: ""
        }

        // Start Bluetooth discovery; emits bfs:printer:found events and bfs:printer:discovery:finished
        @JavascriptInterface
        fun startDiscovery(): Boolean {
            if (!ensureBtPermissions()) return false
            val mgr = getSystemService(BLUETOOTH_SERVICE) as BluetoothManager
            val adapter = mgr.adapter
            if (adapter == null || !adapter.isEnabled) return false
            if (!hasBtScanPermission()) return false
            try {
                if (adapter.isDiscovering) {
                    try { adapter.cancelDiscovery() } catch (_: Throwable) {}
                }
                registerBtReceiver()
                discoveryActive = adapter.startDiscovery()
                return discoveryActive
            } catch (_: SecurityException) {
                return false
            } catch (_: Throwable) {
                return false
            }
        }

        // Stop Bluetooth discovery
        @JavascriptInterface
        fun stopDiscovery() {
            try {
                val mgr = getSystemService(BLUETOOTH_SERVICE) as BluetoothManager
                val adapter = mgr.adapter
                if (adapter != null && adapter.isDiscovering) {
                    try { adapter.cancelDiscovery() } catch (_: Throwable) {}
                }
            } catch (_: Throwable) {}
            discoveryActive = false
        }

        // Attempt to pair (bond) with a device by MAC address
        @JavascriptInterface
        fun pair(address: String?): Boolean {
            if (!ensureBtPermissions()) return false
            val mac = address?.trim().orEmpty()
            if (mac.isBlank()) return false
            return try {
                val mgr = getSystemService(BLUETOOTH_SERVICE) as BluetoothManager
                val adapter = mgr.adapter
                if (adapter == null || !adapter.isEnabled) return false
                if (!hasBtConnectPermission()) return false
                val dev = try { adapter.getRemoteDevice(mac) } catch (_: SecurityException) { return false }
                // Kick off bond; user may see a pairing dialog
                try { dev.createBond() } catch (_: SecurityException) { return false }
                true
            } catch (_: Throwable) {
                false
            }
        }
    }

    private fun startSumUpCheckout(p: PendingPayment) {
        pendingPayment = p
        val payment = try {
            SumUpPayment.builder()
                .total(BigDecimal(p.amountCents).divide(BigDecimal(100)))
                .currency(SumUpPayment.Currency.valueOf(p.currency))
                .title("POS ${p.reference}")
                .foreignTransactionId(p.reference)
                .skipSuccessScreen()
                .build()
        } catch (t: Throwable) {
            sendEvent(
                "bfs:sumup:result",
                JSONObject().put("success", false).put("error", "invalid_currency").put("correlationId", p.reference)
            )
            return
        }

        if (BuildConfig.SUMUP_AFFILIATE_KEY.isBlank()) {
            sendEvent(
                "bfs:sumup:result",
                JSONObject().put("success", false).put("error", "missing_affiliate_key").put("correlationId", p.reference)
            )
            return
        }

        if (SumUpAPI.isLoggedIn()) {
            SumUpAPI.checkout(this, payment, REQ_CHECKOUT)
        } else {
            val login = SumUpLogin.builder(BuildConfig.SUMUP_AFFILIATE_KEY).build()
            SumUpAPI.openLoginActivity(this, login, REQ_LOGIN)
        }
    }

    // SumUpPayment.Builder now uses .skipSuccessScreen() explicitly per SDK docs

    @Deprecated("SumUp SDK uses legacy result callback")
    override fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?) {
        super.onActivityResult(requestCode, resultCode, data)
        val bundle = data?.extras ?: return
        when (requestCode) {
            REQ_LOGIN -> {
                val rc = bundle.getInt(SumUpAPI.Response.RESULT_CODE)
                if (rc == SumUpAPI.Response.ResultCode.SUCCESSFUL) {
                    pendingPayment?.let { startSumUpCheckout(it) }
                } else {
                    val ref = pendingPayment?.reference ?: ""
                    sendEvent(
                        "bfs:sumup:result",
                        JSONObject().put("success", false).put("error", "login_failed").put("correlationId", ref)
                    )
                    pendingPayment = null
                }
            }
            REQ_CHECKOUT -> {
                val ref = pendingPayment?.reference ?: ""
                val rc = bundle.getInt(SumUpAPI.Response.RESULT_CODE)
                if (rc == SumUpAPI.Response.ResultCode.SUCCESSFUL) {
                    val txCode = bundle.getString(SumUpAPI.Response.TX_CODE)

                    // Read TX_INFO parcelable via reflection for broad SDK compatibility
                    val txJson = JSONObject()
                    try {
                        val parcel = try { data?.getParcelableExtra<android.os.Parcelable>(SumUpAPI.Response.TX_INFO) } catch (_: Throwable) { null }
                        if (parcel != null) {
                            fun getString(method: String): String? = try { parcel.javaClass.getMethod(method).invoke(parcel)?.toString() } catch (_: Throwable) { null }
                            // These getters correspond to SDK’s TransactionInfo fields
                            getString("getLastFourDigits")?.let { txJson.put("cardLast4", it) }
                            getString("getCardType")?.let { txJson.put("cardType", it) }
                            getString("getEntryMode")?.let { txJson.put("entryMode", it) }
                            getString("getCurrency")?.let { txJson.put("currency", it) }
                            getString("getTransactionCode")?.let { code -> txJson.put("txId", code) }
                            getString("getStatus")?.let { s -> txJson.put("status", s) }
                        }
                    } catch (_: Throwable) { /* ignore */ }

                    // Always include txId (from bundle) and correlation id
                    if (!txJson.has("txId")) txJson.put("txId", txCode ?: "")
                    txJson.put("correlationId", ref)
                    // Fill amount/currency from the pending payment if still missing
                    pendingPayment?.let {
                        if (!txJson.has("amountCents")) txJson.put("amountCents", it.amountCents)
                        if (!txJson.has("currency")) txJson.put("currency", it.currency)
                    }
                    // Mark overall status
                    txJson.put("success", true)

                    // Keep for later augmentation during print
                    if (ref.isNotBlank()) txByCorrelation[ref] = txJson
                    lastTx = txJson

                    sendEvent(
                        "bfs:sumup:result",
                        txJson
                    )
                } else {
                    val msg = bundle.getString(SumUpAPI.Response.MESSAGE) ?: "payment_failed"
                    sendEvent(
                        "bfs:sumup:result",
                        JSONObject().put("success", false).put("error", msg).put("correlationId", ref)
                    )
                }
                pendingPayment = null
            }
        }
    }

    // Note: We only surface non-sensitive card details (brand, last4, entry mode)

    private fun sendEvent(name: String, detail: JSONObject) {
        val js = (
            "(function(){try{var d=" + JSONObject.quote(detail.toString()) +
                ";var o=JSON.parse(d);window.dispatchEvent(new CustomEvent('" + name + "',{detail:o}));}catch(e){console.error(e);}})();"
            )
        runOnUiThread { webView.evaluateJavascript(js, null) }
    }

    private fun ensureBtPermissions(): Boolean {
        val needs = mutableListOf<String>()
        if (Build.VERSION.SDK_INT >= 31) {
            if (ContextCompat.checkSelfPermission(this, Manifest.permission.BLUETOOTH_CONNECT) != PackageManager.PERMISSION_GRANTED) {
                needs.add(Manifest.permission.BLUETOOTH_CONNECT)
            }
            if (ContextCompat.checkSelfPermission(this, Manifest.permission.BLUETOOTH_SCAN) != PackageManager.PERMISSION_GRANTED) {
                needs.add(Manifest.permission.BLUETOOTH_SCAN)
            }
        }
        return if (needs.isNotEmpty()) {
            ActivityCompat.requestPermissions(this, needs.toTypedArray(), 2001)
            false
        } else true
    }

    private fun hasBtConnectPermission(): Boolean {
        return if (Build.VERSION.SDK_INT >= 31) {
            ContextCompat.checkSelfPermission(this, Manifest.permission.BLUETOOTH_CONNECT) == PackageManager.PERMISSION_GRANTED
        } else true
    }

    private fun hasBtScanPermission(): Boolean {
        return if (Build.VERSION.SDK_INT >= 31) {
            ContextCompat.checkSelfPermission(this, Manifest.permission.BLUETOOTH_SCAN) == PackageManager.PERMISSION_GRANTED
        } else true
    }

    // Very small HTML print helper via hidden WebView
    private fun printWithSystem(htmlContent: String, onResult: (success: Boolean, error: String?) -> Unit) {
        val pv = WebView(this)
        pv.settings.javaScriptEnabled = false
        pv.loadDataWithBaseURL(null, htmlContent, "text/html", "UTF-8", null)
        pv.post {
            try {
                val printMgr = getSystemService(android.content.Context.PRINT_SERVICE) as android.print.PrintManager
                val adapter = pv.createPrintDocumentAdapter("POS_Receipt")
                val job = printMgr.print("POS_Receipt", adapter, android.print.PrintAttributes.Builder().build())

                // Poll job state until it finishes one way or another
                val handler = android.os.Handler(android.os.Looper.getMainLooper())
                fun check() {
                    when {
                        job.isCompleted -> onResult(true, null)
                        job.isFailed -> onResult(false, "print_failed")
                        job.isCancelled -> onResult(false, "print_cancelled")
                        else -> handler.postDelayed({ check() }, 400)
                    }
                }
                check()
            } catch (t: Throwable) {
                onResult(false, t.message ?: "print_failed")
            }
        }
    }

    // Minimal ESC/POS text printing using DantSu library; returns success
    private fun printWithEscPos(content: String, preferredAddress: String? = null): Boolean {
        if (!ensureBtPermissions()) return false
        return try {
            val connection = run {
                val addr = preferredAddress ?: loadSelectedPrinter()
                if (!addr.isNullOrBlank()) {
                    val mgr = getSystemService(BLUETOOTH_SERVICE) as BluetoothManager
                    val adapter = mgr.adapter
                    if (adapter == null || !adapter.isEnabled) throw Exception("bt_disabled")
                    if (!hasBtConnectPermission()) throw Exception("missing_bt_permission")
                    val dev = try { adapter.getRemoteDevice(addr) } catch (_: SecurityException) { throw Exception("bt_connect_permission_denied") }
                    com.dantsu.escposprinter.connection.bluetooth.BluetoothConnection(dev)
                } else {
                    com.dantsu.escposprinter.connection.bluetooth.BluetoothPrintersConnections.selectFirstPaired()
                        ?: throw Exception("no_paired_printer")
                }
            }
            // If the paired device looks like a Phomemo T02 (non-ESC/POS), use raw raster path
            val devName = if (hasBtConnectPermission()) {
                try { connection.device?.name } catch (_: SecurityException) { null } catch (_: Throwable) { null }
            } else null
            val devAddr = if (hasBtConnectPermission()) {
                try { connection.device?.address } catch (_: SecurityException) { null } catch (_: Throwable) { null }
            } else null
            if ((devName?.contains("T02", ignoreCase = true) == true || devName?.contains("Phomemo", ignoreCase = true) == true) && !devAddr.isNullOrBlank()) {
                Log.d("POS_PRINT", "Detected Phomemo-like printer: $devName ($devAddr), using raw raster protocol")
                try { connection.disconnect() } catch (_: Throwable) {}
                return printWithPhomemo(devAddr, content)
            }

            // Try explicit connect to surface errors early for ESC/POS printers
            try {
                connection.connect()
                try {
                    val name = if (hasBtConnectPermission()) {
                        try { connection.device.name } catch (_: SecurityException) { "Bluetooth Printer" } catch (_: Throwable) { "Bluetooth Printer" }
                    } else "Bluetooth Printer"
                    Log.d("POS_PRINT", "Connected to printer: $name")
                } catch (_: Throwable) {}
            } catch (t: Throwable) {
                Log.e("POS_PRINT", "Bluetooth connect failed", t)
                throw Exception(t.message ?: "bt_connect_failed")
            }
            val printer = com.dantsu.escposprinter.EscPosPrinter(connection, 203, 48f, 32)
            // Render a simple receipt from JSON content
            val obj = try { JSONObject(content) } catch (_: Exception) { JSONObject() }
            val totalCents = obj.optLong("totalCents", -1)
            val orderId = obj.optString("orderId", "-")
            val method = obj.optString("method", "card")
            val currency = obj.optString("currency", "CHF")
            val last4 = obj.optString("cardLast4", "")
            val brand = obj.optString("cardType", "")
            val entry = obj.optString("entryMode", "")
            val txId = obj.optString("txId", obj.optString("transactionCode", ""))
            val address = "Industriestrasse 5, 3600 Thun"
            val orderTs = obj.optLong("orderTimestamp", -1L)
            val dateStr = try {
                val whenMs = if (orderTs > 0) orderTs else System.currentTimeMillis()
                val fmt = java.text.SimpleDateFormat("dd.MM.yyyy HH:mm", java.util.Locale.GERMANY)
                fmt.format(java.util.Date(whenMs))
            } catch (_: Throwable) { "" }
            val lines = StringBuilder()
            lines.append("[C]<b>BlessThun Food</b>\n")
            lines.append("[C]$address\n")
            if (dateStr.isNotEmpty()) lines.append("[C]$dateStr\n")
            // 1) Items
            run {
                val items = obj.optJSONArray("items")
                if (items != null) {
                    for (i in 0 until items.length()) {
                        val it = items.optJSONObject(i)
                        val qty = it?.optInt("quantity", 1) ?: 1
                        val title = it?.optString("title", "Item") ?: "Item"
                        val unit = it?.optLong("unitPriceCents", 0L) ?: 0L
                        val lineTotal = unit * qty
                        val priceText = String.format("%s %.2f", currency, lineTotal / 100.0)
                        lines.append("${qty}x ${title} - ${priceText}\n")
                    }
                }
            }
            // 2) Border
            lines.append("[C]------------------------------\n")
            // 3) Cash received/change
            if (method.equals("cash", ignoreCase = true)) {
                val rec = obj.optLong("amountReceivedCents", -1L)
                val chg = obj.optLong("changeCents", -1L)
                if (rec >= 0) {
                    val recStr = String.format("%s %.2f", currency, rec.toDouble() / 100.0)
                    lines.append("Erhalten: $recStr\n")
                }
                if (chg >= 0) {
                    val chgStr = String.format("%s %.2f", currency, chg.toDouble() / 100.0)
                    lines.append("Rückgeld: $chgStr\n")
                }
            }
            // 4) Border (only if we printed cash details)
            val hasCashDetails = method.equals("cash", ignoreCase = true) &&
                (obj.optLong("amountReceivedCents", -1L) >= 0 || obj.optLong("changeCents", -1L) >= 0)
            if (hasCashDetails) {
                lines.append("[C]------------------------------\n")
            }
            // 5) Total
            if (totalCents >= 0) {
                val amt = String.format("%s %.2f", currency, totalCents.toDouble() / 100.0)
                lines.append("Total: $amt\n")
            }
            // 6) Whitespace divider
            lines.append("\n")
            // 7) Method (and card details for non-cash)
            lines.append("Zahlart: ${method}\n")
            if (!method.equals("cash", ignoreCase = true)) {
                if (brand.isNotEmpty() || last4.isNotEmpty() || entry.isNotEmpty()) {
                    val dots = if (last4.isNotEmpty()) " •••• $last4" else ""
                    val entryLabel = if (entry.isNotEmpty()) {
                        val em = entry.lowercase()
                        when {
                            em.contains("contact") -> "Kontaktlos"
                            em.contains("chip") -> "Chip"
                            em.contains("swipe") || em.contains("magnet") -> "Magnetstreifen"
                            else -> entry.uppercase()
                        }
                    } else ""
                    val cardLine = when {
                        brand.isNotEmpty() || last4.isNotEmpty() -> "Karte: ${brand}$dots" + (if (entryLabel.isNotEmpty()) " ($entryLabel)" else "")
                        entryLabel.isNotEmpty() -> "Karte: $entryLabel"
                        else -> null
                    }
                    cardLine?.let { lines.append(it + "\n") }
                }
                if (txId.isNotEmpty()) {
                    lines.append("Transaktion: $txId\n")
                }
            }
            // 8) Thank you
            lines.append("Bestellung: ${orderId}\n")
            lines.append("\n")
            lines.append("[C]Danke!\n\n")
            printer.printFormattedText(lines.toString())
            try { connection.disconnect() } catch (_: Throwable) {}
            true
        } catch (e: Exception) {
            e.printStackTrace()
            // Only show toast for unexpected errors, not for "no_paired_printer"
            if (e.message != "no_paired_printer") {
                try { Toast.makeText(this, "Drucken fehlgeschlagen: ${e.message}", Toast.LENGTH_SHORT).show() } catch (_: Throwable) {}
            }
            false
        }
    }

    // Raw raster printing for Phomemo T02-like printers using SPP socket
    private fun printWithPhomemo(macAddress: String, content: String): Boolean {
        if (!ensureBtPermissions()) return false
        return try {
            val mgr = getSystemService(BLUETOOTH_SERVICE) as BluetoothManager
            val adapter = mgr.adapter
            if (adapter == null || !adapter.isEnabled) throw Exception("bt_disabled")
            if (!hasBtConnectPermission()) throw Exception("missing_bt_permission")
            val dev = try {
                adapter.getRemoteDevice(macAddress)
            } catch (_: SecurityException) {
                throw Exception("bt_connect_permission_denied")
            }
            if (hasBtScanPermission()) {
                try { adapter.cancelDiscovery() } catch (_: SecurityException) {}
            }

            val sock = try {
                dev.createInsecureRfcommSocketToServiceRecord(UUID.fromString("00001101-0000-1000-8000-00805F9B34FB"))
            } catch (_: SecurityException) {
                throw Exception("bt_connect_permission_denied")
            }
            try {
                if (!hasBtConnectPermission()) throw Exception("missing_bt_permission")
                try {
                    sock.connect()
                } catch (_: SecurityException) {
                    throw Exception("bt_connect_permission_denied")
                }
                val name = if (hasBtConnectPermission()) {
                    try { dev.name } catch (_: SecurityException) { macAddress } catch (_: Throwable) { macAddress }
                } else macAddress
                Log.d("POS_PRINT", "Connected (raw) to printer: $name")

                val obj = try { JSONObject(content) } catch (_: Exception) { JSONObject() }
                val bmp = renderReceiptBitmap(obj)

                val out = sock.outputStream
                val wBytes = bmp.width / 8
                val hDots = bmp.height
                val lineBuf = ByteArray(wBytes)

                // Header/init + vendor-specific raster start
                out.write(byteArrayOf(
                    0x1B, 0x40,             // ESC @ init
                    0x1B, 0x61, 0x00,       // ESC a 0 left
                    0x1F, 0x11, 0x02, 0x04  // vendor: start raster
                ))

                var row = 0
                while (row < hDots) {
                    val blockLines = kotlin.math.min(255, hDots - row)
                    // GS v 0 m=0 xL xH yL yH
                    out.write(byteArrayOf(
                        0x1D, 0x76, 0x30, 0x00,
                        (wBytes and 0xFF).toByte(), 0x00,
                        (blockLines and 0xFF).toByte(), (blockLines shr 8).toByte()
                    ))

                    repeat(blockLines) {
                        java.util.Arrays.fill(lineBuf, 0.toByte())
                        var byte = 0
                        var bit = 7
                        var col = 0
                        while (col < bmp.width) {
                            val px = bmp.getPixel(col, row)
                            val lum = Color.red(px) // grayscale
                            if (lum < 128) byte = byte or (1 shl bit)
                            if (bit == 0) {
                                lineBuf[col / 8] = byte.toByte()
                                byte = 0
                                bit = 7
                            } else {
                                bit--
                            }
                            col++
                        }
                        out.write(lineBuf)
                        row++
                    }
                }

                // Footer/commit — add a bit more feed so the final line isn't clipped
                out.write(byteArrayOf(
                    0x1B, 0x64, 0x02,  // ESC d 2  feed 2 lines
                    0x1B, 0x64, 0x02,  // feed 2 more
                    0x1B, 0x64, 0x02,  // feed 2 more (total ~6)
                    0x1F, 0x11, 0x08,  // commit/print sequence
                    0x1F, 0x11, 0x0E,
                    0x1F, 0x11, 0x07,
                    0x1F, 0x11, 0x09
                ))
                out.flush()

                try { Thread.sleep(600) } catch (_: Throwable) {}
                try { out.close() } catch (_: Throwable) {}
                true
            } finally {
                try { sock.close() } catch (_: Throwable) {}
            }
        } catch (t: Throwable) {
            Log.e("POS_PRINT", "Phomemo print failed", t)
            try { Toast.makeText(this, "Drucken fehlgeschlagen: ${t.message}", Toast.LENGTH_SHORT).show() } catch (_: Throwable) {}
            false
        }
    }

    private fun renderReceiptBitmap(obj: JSONObject): Bitmap {
        val width = 384
        val padding = 10
        val headerSize = 24f
        val fontSize = 20f
        val smallSize = 18f
        val lineSpacing = 8
        val headerTopPad = 12
        val headerBottomPad = 6
        val bottomPad = 56
        val qrSize = 320 // make QR code larger (was 260)

        val paint = Paint(Paint.ANTI_ALIAS_FLAG).apply { color = Color.BLACK; textSize = fontSize }
        val bold = Paint(Paint.ANTI_ALIAS_FLAG).apply { color = Color.BLACK; textSize = headerSize; isFakeBoldText = true }
        val small = Paint(Paint.ANTI_ALIAS_FLAG).apply { color = Color.BLACK; textSize = smallSize }

        fun measureHeight(): Int {
            var h = 0
            val fmHeader = Paint.FontMetrics(); bold.getFontMetrics(fmHeader)
            val fm = Paint.FontMetrics(); paint.getFontMetrics(fm)
            val fmSmall = Paint.FontMetrics(); small.getFontMetrics(fmSmall)
            // Header (top pad + header + bottom pad)
            h += headerTopPad + (fmHeader.bottom - fmHeader.top).toInt() + headerBottomPad
            // Address and datetime lines (small)
            h += (fmSmall.bottom - fmSmall.top).toInt() + lineSpacing // address
            h += (fmSmall.bottom - fmSmall.top).toInt() + lineSpacing // date/time
            // QR (optional)
            val hasQr = obj.optString("pickupQr", "").isNotEmpty()
            if (hasQr) h += qrSize + lineSpacing
            // Items
            val items = obj.optJSONArray("items")
            if (items != null) {
                for (i in 0 until items.length()) {
                    h += (fm.bottom - fm.top).toInt() + lineSpacing
                    val it = items.optJSONObject(i)
                    val cfg = it?.optJSONArray("configuration")
                    if (cfg != null) {
                        for (j in 0 until cfg.length()) h += (fmSmall.bottom - fmSmall.top).toInt() + (lineSpacing / 2)
                    }
                }
            }
            // Divider (after items)
            h += (fm.bottom - fm.top).toInt() + lineSpacing
            // Cash: received + change (up to two small lines)
            val methodForHeight = obj.optString("method", "card")
            if (methodForHeight.equals("cash", ignoreCase = true)) {
                val rec = obj.optLong("amountReceivedCents", -1L)
                val chg = obj.optLong("changeCents", -1L)
                if (rec >= 0) h += (fmSmall.bottom - fmSmall.top).toInt() + lineSpacing
                if (chg >= 0) h += (fmSmall.bottom - fmSmall.top).toInt() + lineSpacing
            }
            // Second divider only if cash details are present
            if (methodForHeight.equals("cash", ignoreCase = true)) {
                val rec = obj.optLong("amountReceivedCents", -1L)
                val chg = obj.optLong("changeCents", -1L)
                if (rec >= 0 || chg >= 0) {
                    h += (fm.bottom - fm.top).toInt() + lineSpacing
                }
            }
            // Total
            h += (fm.bottom - fm.top).toInt() + lineSpacing
            // Whitespace divider
            h += (fm.bottom - fm.top).toInt() / 2 + lineSpacing
            // Payment line
            h += (fm.bottom - fm.top).toInt() + lineSpacing
            // Optional card + tx lines (small)
            val hasCardDetails = obj.optString("cardType", "").isNotEmpty() || obj.optString("cardLast4", "").isNotEmpty() || obj.optString("entryMode", "").isNotEmpty()
            if (hasCardDetails) h += (fmSmall.bottom - fmSmall.top).toInt() + lineSpacing
            if (obj.optString("txId", obj.optString("transactionCode", "")).isNotEmpty()) h += (fmSmall.bottom - fmSmall.top).toInt() + lineSpacing
            // Order ID (small) under method
            h += (fmSmall.bottom - fmSmall.top).toInt() + lineSpacing
            // Whitespace divider
            h += (fm.bottom - fm.top).toInt() / 2 + lineSpacing
            // Danke + bottom pad
            h += (fm.bottom - fm.top).toInt() + bottomPad
            return h.coerceAtLeast(200)
        }

        val bmp = Bitmap.createBitmap(width, measureHeight(), Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bmp)
        canvas.drawColor(Color.WHITE)

        var y: Float
        run {
            val fm = Paint.FontMetrics(); bold.getFontMetrics(fm)
            y = headerTopPad - fm.top
            val title = "BlessThun Food"
            val tw = bold.measureText(title)
            canvas.drawText(title, (width - tw) / 2f, y, bold)
            y += (fm.bottom - fm.top) + headerBottomPad
        }

        // Address (centered, small)
        run {
            val fmSmall = Paint.FontMetrics(); small.getFontMetrics(fmSmall)
            val addr = "Industriestrasse 5, 3600 Thun"
            val aw = small.measureText(addr)
            canvas.drawText(addr, (width - aw) / 2f, y - fmSmall.top, small)
            y += (fmSmall.bottom - fmSmall.top) + lineSpacing
        }

        // Date/time of order (centered, small). Prefer orderTimestamp if provided.
        run {
            val fmSmall = Paint.FontMetrics(); small.getFontMetrics(fmSmall)
            val orderTs = obj.optLong("orderTimestamp", -1L)
            val dateStr = try {
                val whenMs = if (orderTs > 0) orderTs else System.currentTimeMillis()
                val fmt = java.text.SimpleDateFormat("dd.MM.yyyy HH:mm", java.util.Locale.GERMANY)
                fmt.format(java.util.Date(whenMs))
            } catch (_: Throwable) { "" }
            if (dateStr.isNotEmpty()) {
                val dw = small.measureText(dateStr)
                canvas.drawText(dateStr, (width - dw) / 2f, y - fmSmall.top, small)
                y += (fmSmall.bottom - fmSmall.top) + lineSpacing
            }
        }

        // QR code if provided
        val pickupQr = obj.optString("pickupQr", "")
        if (pickupQr.isNotEmpty()) {
            val size = qrSize
            try {
                val matrix: BitMatrix = QRCodeWriter().encode(pickupQr, BarcodeFormat.QR_CODE, size, size)
                val qrBmp = Bitmap.createBitmap(size, size, Bitmap.Config.RGB_565)
                for (x in 0 until size) {
                    for (yy in 0 until size) {
                        qrBmp.setPixel(x, yy, if (matrix.get(x, yy)) Color.BLACK else Color.WHITE)
                    }
                }
                val left = (width - size) / 2f
                canvas.drawBitmap(qrBmp, left, y, null)
                y += size + lineSpacing
            } catch (t: Throwable) {
                // ignore QR failures, continue text
            }
        }

        // Items list
        val items = obj.optJSONArray("items")
        if (items != null) {
            val fm = Paint.FontMetrics(); paint.getFontMetrics(fm)
            val fmSmall = Paint.FontMetrics(); small.getFontMetrics(fmSmall)
            for (i in 0 until items.length()) {
                val it = items.optJSONObject(i)
                val qty = it?.optInt("quantity", 1) ?: 1
                val title = it?.optString("title", "Item") ?: "Item"
                val unit = it?.optLong("unitPriceCents", 0L) ?: 0L
                val lineTotal = unit * qty
                val priceText = String.format("CHF %.2f", lineTotal / 100.0)
                val priceW = paint.measureText(priceText)
                val leftX = padding.toFloat()
                val rightX = width - padding - priceW

                // Ellipsize title to fit
                val maxTitleWidth = rightX - leftX - 12
                val titleText = ellipsize("${qty}x ${title}", paint, maxTitleWidth)
                canvas.drawText(titleText, leftX, y - fm.top, paint)
                canvas.drawText(priceText, rightX, y - fm.top, paint)
                y += (fm.bottom - fm.top) + lineSpacing

                // Configuration lines
                val cfg = it?.optJSONArray("configuration")
                if (cfg != null) {
                    for (j in 0 until cfg.length()) {
                        val c = cfg.optJSONObject(j)
                        val slotName = c?.optString("slot", "") ?: ""
                        val choiceName = c?.optString("choice", "") ?: ""
                        val line = "- $slotName: $choiceName"
                        canvas.drawText(ellipsize(line, small, (width - padding * 2).toFloat()), leftX + 14f, y - fmSmall.top, small)
                        y += (fmSmall.bottom - fmSmall.top) + (lineSpacing / 2)
                    }
                }
            }
        }

        // Divider after items
        run {
            val fm = Paint.FontMetrics(); paint.getFontMetrics(fm)
            canvas.drawText("------------------------------", padding.toFloat(), y - fm.top, paint)
            y += (fm.bottom - fm.top) + lineSpacing
        }

        // Totals and meta
        val totalCents = obj.optLong("totalCents", -1)
        val currency = obj.optString("currency", "CHF")
        // Read method now to position cash lines earlier
        val method = obj.optString("method", "card")
        val orderId = obj.optString("orderId", "-")

        // 3) Cash details after first divider (only for cash)
        if (method.equals("cash", ignoreCase = true)) {
            val fmSmall = Paint.FontMetrics(); small.getFontMetrics(fmSmall)
            val rec = obj.optLong("amountReceivedCents", -1L)
            val chg = obj.optLong("changeCents", -1L)
            if (rec >= 0) {
                val recStr = String.format("Erhalten: %s %.2f", currency, rec / 100.0)
                val line = ellipsize(recStr, small, (width - padding * 2).toFloat())
                canvas.drawText(line, padding.toFloat(), y - fmSmall.top, small)
                y += (fmSmall.bottom - fmSmall.top) + lineSpacing
            }
            if (chg >= 0) {
                val chgStr = String.format("Rückgeld: %s %.2f", currency, chg / 100.0)
                val line = ellipsize(chgStr, small, (width - padding * 2).toFloat())
                canvas.drawText(line, padding.toFloat(), y - fmSmall.top, small)
                y += (fmSmall.bottom - fmSmall.top) + lineSpacing
            }
        }

        // 4) Second divider only if we actually printed cash details
        run {
            val rec = obj.optLong("amountReceivedCents", -1L)
            val chg = obj.optLong("changeCents", -1L)
            if (method.equals("cash", ignoreCase = true) && (rec >= 0 || chg >= 0)) {
                val fm = Paint.FontMetrics(); paint.getFontMetrics(fm)
                canvas.drawText("------------------------------", padding.toFloat(), y - fm.top, paint)
                y += (fm.bottom - fm.top) + lineSpacing
            }
        }

        // 5) Total
        if (totalCents >= 0) {
            val fm = Paint.FontMetrics(); paint.getFontMetrics(fm)
            val t = String.format("Total: %s %.2f", currency, totalCents / 100.0)
            canvas.drawText(t, padding.toFloat(), y - fm.top, paint)
            y += (fm.bottom - fm.top) + lineSpacing
        }

        // 6) Whitespace divider
        run {
            val fm = Paint.FontMetrics(); paint.getFontMetrics(fm)
            y += (fm.bottom - fm.top) / 2 + lineSpacing
        }
        run {
            // Payment line with capitalized method
            val fm = Paint.FontMetrics(); paint.getFontMetrics(fm)
            val methodPretty = if (method.isNotEmpty()) method.substring(0,1).uppercase() + method.substring(1) else method
            val payLine = "Zahlart: $methodPretty"
            val payText = ellipsize(payLine, paint, (width - padding * 2).toFloat())
            canvas.drawText(payText, padding.toFloat(), y - fm.top, paint)
            y += (fm.bottom - fm.top) + lineSpacing
        }

        // Cash details already printed above after the first divider

        // Card details (small font)
        run {
            val fmSmall = Paint.FontMetrics(); small.getFontMetrics(fmSmall)
            val brand = obj.optString("cardType", "")
            val last4 = obj.optString("cardLast4", "")
            val entry = obj.optString("entryMode", "")
            if (brand.isNotEmpty() || last4.isNotEmpty() || entry.isNotEmpty()) {
                val dots = if (last4.isNotEmpty()) " •••• $last4" else ""
                val entryLabel = if (entry.isNotEmpty()) {
                    val em = entry.lowercase()
                    when {
                        em.contains("contact") -> "Kontaktlos"
                        em.contains("chip") -> "Chip"
                        em.contains("swipe") || em.contains("magnet") -> "Magnetstreifen"
                        else -> entry.uppercase()
                    }
                } else ""
                val text = when {
                    brand.isNotEmpty() || last4.isNotEmpty() -> "Karte: ${brand}$dots" + (if (entryLabel.isNotEmpty()) " ($entryLabel)" else "")
                    entryLabel.isNotEmpty() -> "Karte: $entryLabel"
                    else -> null
                }
                text?.let {
                    val cardLine = ellipsize(it, small, (width - padding * 2).toFloat())
                    canvas.drawText(cardLine, padding.toFloat(), y - fmSmall.top, small)
                    y += (fmSmall.bottom - fmSmall.top) + lineSpacing
                }
            }
            val txId = obj.optString("txId", obj.optString("transactionCode", ""))
            if (txId.isNotEmpty()) {
                val txLine = ellipsize("Transaktion: $txId", small, (width - padding * 2).toFloat())
                canvas.drawText(txLine, padding.toFloat(), y - fmSmall.top, small)
                y += (fmSmall.bottom - fmSmall.top) + lineSpacing
            }
            // no explicit status line — success is implied
        }

        // Order ID under method
        run {
            val fmSmall = Paint.FontMetrics(); small.getFontMetrics(fmSmall)
            val ordLine = ellipsize("Bestellung: $orderId", small, (width - padding * 2).toFloat())
            canvas.drawText(ordLine, padding.toFloat(), y - fmSmall.top, small)
            y += (fmSmall.bottom - fmSmall.top) + lineSpacing
        }
        // Whitespace divider before thank you
        run {
            val fm = Paint.FontMetrics(); paint.getFontMetrics(fm)
            y += (fm.bottom - fm.top) / 2 + lineSpacing
        }

        // Danke (with extra bottom padding to avoid clipping)
        run {
            val fm = Paint.FontMetrics(); paint.getFontMetrics(fm)
            canvas.drawText("Danke!", padding.toFloat(), y - fm.top, paint)
            // leave some bottom padding to avoid clipping on printers
            y += bottomPad
        }

        return bmp
    }

    private fun ellipsize(text: String, paint: Paint, maxWidth: Float): String {
        if (paint.measureText(text) <= maxWidth) return text
        var s = text
        val ell = "…"
        var w = paint.measureText(s + ell)
        while (s.isNotEmpty() && w > maxWidth) {
            s = s.substring(0, s.length - 1)
            w = paint.measureText(s + ell)
        }
        return s + ell
    }
}
