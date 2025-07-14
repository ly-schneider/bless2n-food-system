package ch.leys.bless2n

import android.app.Application
import com.sumup.merchant.reader.api.SumUpState

class ShopApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        // one-time SDK bootstrap
        SumUpState.init(this)
    }
}