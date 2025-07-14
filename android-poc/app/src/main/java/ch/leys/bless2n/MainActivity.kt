package ch.leys.bless2n

import android.Manifest
import android.bluetooth.BluetoothManager
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.Color
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.compose.setContent
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
import ch.leys.bless2n.ui.theme.AndroidpocTheme
import com.sumup.merchant.reader.api.*
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import java.math.BigDecimal
import java.util.UUID
import kotlinx.coroutines.delay
import com.google.zxing.BarcodeFormat
import com.google.zxing.qrcode.QRCodeWriter
import com.google.zxing.common.BitMatrix
import com.journeyapps.barcodescanner.ScanContract
import com.journeyapps.barcodescanner.ScanOptions
import com.journeyapps.barcodescanner.CaptureActivity
import android.media.ToneGenerator
import android.media.AudioManager

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

        Spacer(Modifier.padding(vertical = 12.dp))
        QRScanButton()
    }
}

@Composable
fun QRScanButton() {
    val context = LocalContext.current
    
    val scanLauncher = rememberLauncherForActivityResult(
        contract = ScanContract()
    ) { result ->
        if (result.contents != null) {
            // Play high-pitched sound
            playSuccessSound()
            // Output QR data to console
            println("QR Code Scanned: ${result.contents}")
            android.util.Log.d("QRScanner", "QR Code Data: ${result.contents}")
        }
    }
    
    val cameraPermissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        if (isGranted) {
            val options = ScanOptions()
            options.setPrompt("Scan a QR code - Tap screen to focus")
            options.setBeepEnabled(false) // We'll use our custom sound
            options.setBarcodeImageEnabled(true)
            options.setOrientationLocked(false)
            options.setDesiredBarcodeFormats(ScanOptions.QR_CODE)
            options.setCameraId(0) // Use back camera
            options.captureActivity = CustomCaptureActivity::class.java
            scanLauncher.launch(options)
        }
    }
    
    Button(
        onClick = {
            when (ContextCompat.checkSelfPermission(context, Manifest.permission.CAMERA)) {
                PackageManager.PERMISSION_GRANTED -> {
                    val options = ScanOptions()
                    options.setPrompt("Scan a QR code - Tap screen to focus")
                    options.setBeepEnabled(false) // We'll use our custom sound
                    options.setBarcodeImageEnabled(true)
                    options.setOrientationLocked(false)
                    options.setDesiredBarcodeFormats(ScanOptions.QR_CODE)
                    options.setCameraId(0) // Use back camera
                    options.captureActivity = CustomCaptureActivity::class.java
                    scanLauncher.launch(options)
                }
                else -> {
                    cameraPermissionLauncher.launch(Manifest.permission.CAMERA)
                }
            }
        },
        colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.tertiary),
        modifier = Modifier.fillMaxWidth()
    ) {
        Text("Scan QR Code")
    }
}

private fun playSuccessSound() {
    try {
        val toneGenerator = ToneGenerator(AudioManager.STREAM_NOTIFICATION, 100)
        toneGenerator.startTone(ToneGenerator.TONE_PROP_BEEP2, 200)
        toneGenerator.release()
    } catch (e: Exception) {
        android.util.Log.e("QRScanner", "Error playing sound", e)
    }
}

@Composable
fun EnsureBtPermission(content: @Composable () -> Unit) {
    val ctx = LocalContext.current
    val launcher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions()
    ) { granted -> if (granted.values.all { it }) Unit }

    LaunchedEffect(Unit) {
        val permissionsNeeded = mutableListOf<String>()
        
        if (ContextCompat.checkSelfPermission(ctx, Manifest.permission.BLUETOOTH_CONNECT) != PackageManager.PERMISSION_GRANTED) {
            permissionsNeeded.add(Manifest.permission.BLUETOOTH_CONNECT)
        }
        if (ContextCompat.checkSelfPermission(ctx, Manifest.permission.BLUETOOTH_SCAN) != PackageManager.PERMISSION_GRANTED) {
            permissionsNeeded.add(Manifest.permission.BLUETOOTH_SCAN)
        }
        
        if (permissionsNeeded.isNotEmpty()) {
            launcher.launch(permissionsNeeded.toTypedArray())
        }
    }
    content()
}

private fun printReceipt(ctx: Context) = CoroutineScope(Dispatchers.IO).launch {

    /* ----------  Bluetooth setup (unchanged)  ---------------------------- */
    val mac = "B0:B0:09:43:96:16"
    val mgr = ctx.getSystemService(Context.BLUETOOTH_SERVICE) as BluetoothManager
    val dev = mgr.adapter.getRemoteDevice(mac)
    if (ContextCompat.checkSelfPermission(ctx, Manifest.permission.BLUETOOTH_CONNECT) != PackageManager.PERMISSION_GRANTED ||
        ContextCompat.checkSelfPermission(ctx, Manifest.permission.BLUETOOTH_SCAN) != PackageManager.PERMISSION_GRANTED) return@launch
    mgr.adapter.cancelDiscovery()

    val spp = UUID.fromString("00001101-0000-1000-8000-00805F9B34FB")        // SPP
    val sock = dev.createInsecureRfcommSocketToServiceRecord(spp)            // insecure = no SSP

    try {
        sock.connect()

        /* ---------- 1. build a 384-pixel-wide QR code bitmap ------------ */
        val identificationCode = "000001"
        val qrCodeWriter = QRCodeWriter()
        val bitMatrix: BitMatrix = qrCodeWriter.encode(identificationCode, BarcodeFormat.QR_CODE, 200, 200)
        
        // Create QR code bitmap
        val qrWidth = bitMatrix.width
        val qrHeight = bitMatrix.height
        val qrBmp = Bitmap.createBitmap(qrWidth, qrHeight, Bitmap.Config.RGB_565)
        
        for (x in 0 until qrWidth) {
            for (y in 0 until qrHeight) {
                qrBmp.setPixel(x, y, if (bitMatrix[x, y]) Color.BLACK else Color.WHITE)
            }
        }
        
        // Create a 384-pixel-wide bitmap with the QR code centered
        val bmp = Bitmap.createBitmap(384, 250, Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bmp)
        canvas.drawColor(Color.WHITE)
        
        // Center the QR code horizontally
        val offsetX = (384 - qrWidth) / 2
        val offsetY = 25 // Some top margin
        canvas.drawBitmap(qrBmp, offsetX.toFloat(), offsetY.toFloat(), null)

        /* ---------- 2. convert to ESC/POS raster blocks ------------------ */
        val wBytes = 48                      // 384 px ÷ 8
        val hDots  = bmp.height
        val out    = sock.outputStream

        // --- 2.1 printer header (Phomemo-specific) -----------------------
        out.write(byteArrayOf(
            0x1B,0x40,                     // ESC @  initialise
            0x1B,0x61,0x00,                // ESC a 0  left-justify
            0x1F,0x11,0x02,0x04            // “start raster” magic
        ))

        // --- 2.2 write in ≤255-line chunks (firmware limit) --------------
        var row = 0
        val lineBuf = ByteArray(wBytes)
        while (row < hDots) {
            val blockLines = minOf(255, hDots - row)                         // spec limit
            /* GS v 0 m=0 xL xH yL yH */
            out.write(byteArrayOf(
                0x1D,0x76,0x30,0x00,
                (wBytes and 0xFF).toByte(),0x00,
                (blockLines and 0xFF).toByte(),(blockLines shr 8).toByte()
            ))                                                               // ESC/POS raster cmd

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
            0x1F,0x11,0x08,               // commit / print
            0x1F,0x11,0x0E,
            0x1F,0x11,0x07,
            0x1F,0x11,0x09
        ))
        out.flush()

        /* ---------- 4. give the head time to burn, THEN close ----------- */
        delay(600)                        // ≈0.5 s is enough for 4 text lines
        out.close()                       // explicit close = cleaner than just sock.close()

    } finally {
        sock.close()
    }
}


