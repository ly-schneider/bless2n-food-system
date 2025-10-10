package ch.leys.bless2n

import android.annotation.SuppressLint
import android.content.Intent
import android.os.Bundle
import android.webkit.JavascriptInterface
import android.webkit.WebChromeClient
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.activity.ComponentActivity
import com.sumup.merchant.reader.api.SumUpAPI
import com.sumup.merchant.reader.api.SumUpLogin
import com.sumup.merchant.reader.api.SumUpPayment
import org.json.JSONObject

class PosWebActivity : ComponentActivity() {
    private lateinit var webView: WebView

    companion object {
        private const val REQ_LOGIN = 101
        private const val REQ_CHECKOUT = 102
    }

    @SuppressLint("SetJavaScriptEnabled")
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        webView = WebView(this)
        setContentView(webView)

        val settings = webView.settings
        settings.javaScriptEnabled = true
        settings.domStorageEnabled = true
        settings.mixedContentMode = WebSettings.MIXED_CONTENT_NEVER_ALLOW
        settings.allowFileAccess = false
        settings.allowContentAccess = false
        webView.webChromeClient = WebChromeClient()
        webView.webViewClient = object : WebViewClient() {}

        // Expose minimal bridge under window.PosBridge
        webView.addJavascriptInterface(PosBridge(), "PosBridge")

        // TODO: restrict allowed origins and load only trusted HTTPS content
        val url = intent.getStringExtra("url") ?: "https://app.example/pos"
        webView.loadUrl(url)
    }

    inner class PosBridge {
        @JavascriptInterface
        fun payWithCard(payload: String) {
            // payload: JSON string { amountCents, currency, reference }
            try {
                val obj = JSONObject(payload)
                val amountCents = obj.optLong("amountCents")
                val currency = obj.optString("currency", "CHF")
                val total = amountCents.toBigDecimal().divide(java.math.BigDecimal(100))

                if (!SumUpAPI.isLoggedIn()) {
                    val login = SumUpLogin.builder(BuildConfig.SUMUP_AFFILIATE_KEY).build()
                    SumUpAPI.openLoginActivity(this@PosWebActivity, login, REQ_LOGIN)
                    return
                }

                val payment = SumUpPayment.builder()
                    .total(total)
                    .currency(SumUpPayment.Currency.valueOf(currency))
                    .title("POS sale")
                    .build()
                SumUpAPI.checkout(this@PosWebActivity, payment, REQ_CHECKOUT)
            } catch (_: Throwable) {}
        }

        @JavascriptInterface
        fun print(receiptJson: String) {
            // TODO: Render ESC/POS and send via RFCOMM; see MainActivity.printReceipt
        }
    }

    @Deprecated("SumUp SDK still relies on legacy callback")
    override fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?) {
        super.onActivityResult(requestCode, resultCode, data)
        // TODO: map SDK response back to WebView via evaluateJavascript
    }
}

