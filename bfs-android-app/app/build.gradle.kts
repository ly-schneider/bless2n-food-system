plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
}

android {
    namespace = "ch.leys.bless2n"
    compileSdk = 35

    defaultConfig {
        applicationId = "ch.leys.bless2n"
        // SumUp SDK 5.x requires minSdk >= 26
        minSdk = 26
        targetSdk = 35
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        vectorDrawables { useSupportLibrary = true }

        // SumUp affiliate key from gradle property or environment; avoid hard-coding
        val sumupAffiliateKey = (project.findProperty("sumupAffiliateKey") as String?)
            ?: System.getenv("SUMUP_AFFILIATE_KEY")
            ?: ""
        if (gradle.startParameter.taskNames.any { it.contains("Release", ignoreCase = true) } && sumupAffiliateKey.isBlank()) {
            throw GradleException("SUMUP_AFFILIATE_KEY (or -PsumupAffiliateKey) must be set for release builds")
        }
        buildConfigField("String", "SUMUP_AFFILIATE_KEY", "\"$sumupAffiliateKey\"")

        // Default buildConfig fields so all variants (incl. debug) compile
        buildConfigField("String", "POS_URL", "\"http://127.0.0.1:3000/pos\"")
        buildConfigField("boolean", "DEV_BUILD", "true")
    }

    buildFeatures {
        buildConfig = true
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
            // Configure production POS URL via gradle property or env
            val buildingRelease = gradle.startParameter.taskNames.any { it.contains("Release", ignoreCase = true) }
            val posUrlReleaseProp = (project.findProperty("posUrlRelease") as String?) ?: System.getenv("POS_URL_RELEASE")
            if (buildingRelease && (posUrlReleaseProp.isNullOrBlank())) {
                throw GradleException("POS_URL_RELEASE (or -PposUrlRelease) must be set for release builds")
            }
            val posUrlRelease = posUrlReleaseProp ?: "https://example.com/pos"
            buildConfigField("String", "POS_URL", "\"$posUrlRelease\"")
            buildConfigField("boolean", "DEV_BUILD", "false")
        }

        create("dev") {
            // Mirror debug-like behavior for developer builds
            initWith(getByName("debug"))
            isDebuggable = true
            signingConfig = signingConfigs.getByName("debug")
            applicationIdSuffix = ".dev"
            versionNameSuffix = "-dev"
            // Emulator points to host machine
            val posUrlDev = (project.findProperty("posUrlDev") as String?)
                ?: System.getenv("POS_URL_DEV")
                ?: "http://127.0.0.1:3000/pos"
            buildConfigField("String", "POS_URL", "\"$posUrlDev\"")
            buildConfigField("boolean", "DEV_BUILD", "true")
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }
    kotlinOptions {
        jvmTarget = "11"
    }
    buildFeatures {
        compose = true
    }
}

dependencies {
    implementation("com.sumup:merchant-sdk:5.0.3")
    implementation("com.github.DantSu:ESCPOS-ThermalPrinter-Android:3.3.0")
    implementation("com.google.zxing:core:3.5.2")
    implementation("com.journeyapps:zxing-android-embedded:4.3.0")
    implementation("androidx.webkit:webkit:1.14.0")

    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.activity.compose)
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.ui)
    implementation(libs.androidx.ui.graphics)
    implementation(libs.androidx.ui.tooling.preview)
    implementation(libs.androidx.material3)
    testImplementation(libs.junit)
    androidTestImplementation(libs.androidx.junit)
    androidTestImplementation(libs.androidx.espresso.core)
    androidTestImplementation(platform(libs.androidx.compose.bom))
    androidTestImplementation(libs.androidx.ui.test.junit4)
    debugImplementation(libs.androidx.ui.tooling)
    debugImplementation(libs.androidx.ui.test.manifest)
}
