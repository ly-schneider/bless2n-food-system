package ch.leys.bless2n

import com.journeyapps.barcodescanner.CaptureActivity
import com.journeyapps.barcodescanner.DecoratedBarcodeView
import com.journeyapps.barcodescanner.camera.CameraSettings

class CustomCaptureActivity : CaptureActivity() {
    
    override fun initializeContent(): DecoratedBarcodeView {
        val barcodeView = super.initializeContent()
        
        // Configure camera settings for better focus
        val cameraSettings = CameraSettings()
        cameraSettings.isAutoFocusEnabled = true
        cameraSettings.isContinuousFocusEnabled = true
        cameraSettings.requestedCameraId = 0  // Use back camera
        
        barcodeView.barcodeView.cameraSettings = cameraSettings
        
        // Add tap-to-focus functionality
        barcodeView.setOnClickListener {
            barcodeView.barcodeView.pause()
            barcodeView.barcodeView.resume()
        }
        
        return barcodeView
    }
}