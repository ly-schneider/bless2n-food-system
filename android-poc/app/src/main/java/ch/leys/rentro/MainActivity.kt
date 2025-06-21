package ch.leys.rentro

import android.Manifest
import android.app.Activity
import android.bluetooth.BluetoothDevice
import android.bluetooth.BluetoothManager
import android.bluetooth.BluetoothSocket
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.Color
import android.graphics.Typeface
import android.os.Bundle
import android.text.Layout
import android.text.StaticLayout
import android.text.TextPaint
import androidx.activity.ComponentActivity
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import ch.leys.rentro.ui.theme.AndroidpocTheme
import com.dantsu.escposprinter.EscPosPrinter
import com.dantsu.escposprinter.connection.bluetooth.BluetoothPrintersConnections
import com.sumup.merchant.reader.api.*
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import java.io.ByteArrayOutputStream
import java.math.BigDecimal
import java.util.UUID
import androidx.core.graphics.createBitmap
import androidx.core.graphics.get
import kotlinx.coroutines.delay

class MainActivity : ComponentActivity() {

    companion object {
        internal const val REQ_LOGIN = 1
        private const val REQ_CHECKOUT = 2
    }

    /* ---------- state hoisted at the Activity level ---------- */
    val lastReceipt = mutableStateOf<Bundle?>(null)        // <─ NEW

    private val total = BigDecimal("1.00")
    private val currency = SumUpPayment.Currency.CHF

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        WindowCompat.setDecorFitsSystemWindows(window, false)      /* hide bars */
        WindowInsetsControllerCompat(window, window.decorView).apply {
            hide(WindowInsetsCompat.Type.systemBars())
            systemBarsBehavior =
                WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
        }

        setContent {
            AndroidpocTheme {
                EnsureBtPermission {
                    Scaffold(Modifier.fillMaxSize()) { inner ->
                        PaymentScreen(
                            modifier = Modifier
                                .padding(inner)
                                .fillMaxWidth()
                        )
                    }
                }
            }
        }
    }

    /* ---------- results ---------- */
    @Deprecated("SumUp SDK still relies on legacy callback")
    override fun onActivityResult(req: Int, res: Int, data: Intent?) {
        super.onActivityResult(req, res, data)
        if (res != RESULT_OK || data == null) return

        val bundle = data.extras ?: return
        when (req) {
            REQ_LOGIN -> if (bundle.getInt(SumUpAPI.Response.RESULT_CODE) ==
                SumUpAPI.Response.ResultCode.SUCCESSFUL
            ) startCheckout()

            REQ_CHECKOUT -> if (bundle.getInt(SumUpAPI.Response.RESULT_CODE) ==
                SumUpAPI.Response.ResultCode.SUCCESSFUL
            ) {

                lastReceipt.value = bundle        // <─ store but DON’T print
            }
        }
    }

    /* ---------- launch payment ---------- */
    internal fun startCheckout() {
        val payment = SumUpPayment.builder()
            .total(total)
            .currency(currency)
            .title("Tablet sale")
            .build()
        SumUpAPI.checkout(this, payment, REQ_CHECKOUT)
    }
}

/* ----------  Composables  ---------- */

@Composable
fun PaymentScreen(modifier: Modifier = Modifier) {

    val act = LocalContext.current as MainActivity
    val ctx = LocalContext.current
    val receipt by act.lastReceipt                      // observe state

    Column(modifier) {

        /* primary action: pay */
        Button(
            onClick = {
                if (SumUpAPI.isLoggedIn()) {
                    act.startCheckout()
                } else {
                    val login = SumUpLogin
                        .builder(BuildConfig.SUMUP_AFFILIATE_KEY).build()
                    SumUpAPI.openLoginActivity(act, login, MainActivity.REQ_LOGIN)
                }
            },
            modifier = Modifier.fillMaxWidth()
        ) { Text("Pay with card") }

        Spacer(Modifier.padding(vertical = 12.dp))
        Button(
            onClick = { printReceipt(ctx) },
            colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.secondary),
            modifier = Modifier.fillMaxWidth()
        ) { Text("Print receipt") }
    }
}


@Composable
fun EnsureBtPermission(content: @Composable () -> Unit) {
    val ctx = LocalContext.current
    val launcher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { granted -> if (granted) Unit }

    LaunchedEffect(Unit) {
        if (ContextCompat.checkSelfPermission(
                ctx, Manifest.permission.BLUETOOTH_CONNECT
            ) != PackageManager.PERMISSION_GRANTED
        ) {
            launcher.launch(Manifest.permission.BLUETOOTH_CONNECT)
        }
    }
    content()
}

private fun printReceipt(ctx: Context) = CoroutineScope(Dispatchers.IO).launch {

    /* ----------  Bluetooth setup (unchanged)  ---------------------------- */
    val mac = "B0:B0:09:43:96:16"
    val mgr = ctx.getSystemService(Context.BLUETOOTH_SERVICE) as BluetoothManager
    val dev = mgr.adapter.getRemoteDevice(mac)
    if (ContextCompat.checkSelfPermission(ctx, Manifest.permission.BLUETOOTH_CONNECT)
        != PackageManager.PERMISSION_GRANTED) return@launch
    mgr.adapter.cancelDiscovery()

    val spp = UUID.fromString("00001101-0000-1000-8000-00805F9B34FB")        // SPP UUID:contentReference[oaicite:5]{index=5}
    val sock = dev.createInsecureRfcommSocketToServiceRecord(spp)            // insecure = no SSP:contentReference[oaicite:6]{index=6}

    try {
        sock.connect()

        /* ---------- 1. build a 384-pixel-wide bitmap --------------------- */
        val paint = TextPaint().apply {
            color = Color.BLACK
            textSize = 28f
            typeface = Typeface.MONOSPACE
            isAntiAlias = false          // pure 1-bit pixels
        }
        val msg = """
            MY SHOP
            Tx: 1
            Amount: CHF 1.00
            Thank you!
        """.trimIndent()

        val layout = StaticLayout.Builder.obtain(msg, 0, msg.length, paint, 384)
            .setAlignment(Layout.Alignment.ALIGN_NORMAL)
            .build()                                                         // Android text → bitmap:contentReference[oaicite:7]{index=7}

        // add the font descent so the last baseline isn’t cut
        val extra = paint.fontMetricsInt.descent
        val bmp = Bitmap.createBitmap(384, layout.height + extra, Bitmap.Config.ARGB_8888)
        Canvas(bmp).apply { drawColor(Color.WHITE); layout.draw(this) }

        /* ---------- 2. convert to ESC/POS raster blocks ------------------ */
        val wBytes = 48                      // 384 px ÷ 8
        val hDots  = bmp.height
        val out    = sock.outputStream

        // --- 2.1 printer header (Phomemo-specific) -----------------------
        out.write(byteArrayOf(
            0x1B,0x40,                     // ESC @  initialise
            0x1B,0x61,0x00,                // ESC a 0  left-justify
            0x1F,0x11,0x02,0x04            // “start raster” magic:contentReference[oaicite:8]{index=8}
        ))

        // --- 2.2 write in ≤255-line chunks (firmware limit) --------------
        var row = 0
        val lineBuf = ByteArray(wBytes)
        while (row < hDots) {
            val blockLines = minOf(255, hDots - row)                         // spec limit:contentReference[oaicite:9]{index=9}
            /* GS v 0 m=0 xL xH yL yH */
            out.write(byteArrayOf(
                0x1D,0x76,0x30,0x00,
                (wBytes and 0xFF).toByte(),0x00,
                (blockLines and 0xFF).toByte(),(blockLines shr 8).toByte()
            ))                                                               // ESC/POS raster cmd:contentReference[oaicite:10]{index=10}

            repeat(blockLines) { _ ->
                var byte = 0; var bit = 7; var col = 0
                while (col < 384) {
                    val lum = Color.red(bmp.getPixel(col, row))              // R==G==B
                    if (lum < 128) byte = byte or (1 shl bit)
                    if (bit == 0) { lineBuf[col / 8] = byte.toByte(); byte = 0; bit = 7 }
                    else bit--
                    col++
                }
                out.write(lineBuf)
                row++
            }
        }

        /* ---------- 3. footer & feed ------------------------------------ */
        out.write(byteArrayOf(
            0x1B,0x64,0x02,               // ESC d 2     feed 2 lines
            0x1B,0x64,0x02,               // ESC d 2     again
            0x1F,0x11,0x08,               // commit / print:contentReference[oaicite:11]{index=11}
            0x1F,0x11,0x0E,
            0x1F,0x11,0x07,
            0x1F,0x11,0x09
        ))
        out.flush()

        /* ---------- 4. give the head time to burn, THEN close ----------- */
        delay(600)                        // ≈0.5 s is enough for 4 text lines:contentReference[oaicite:12]{index=12}
        out.close()                       // explicit close = cleaner than just sock.close()

    } finally {
        sock.close()
    }
}


