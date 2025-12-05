import java.util.regex.Pattern
import org.jetbrains.kotlin.gradle.dsl.JvmTarget

fun semverFromTag(): Triple<Int, Int, Int> {
    val raw = (System.getenv("GITHUB_REF_NAME") ?: System.getenv("VERSION_TAG") ?: "")
    val tag = raw.removePrefix("v")
    val m = Pattern.compile("""^(\d+)\.(\d+)\.(\d+)(?:[-+].*)?$""").matcher(tag)
    return if (m.matches()) Triple(
        m.group(1).toInt(),
        m.group(2).toInt(),
        m.group(3).toInt()
    ) else Triple(0, 0, 1)
}

val (maj, min, pat) = semverFromTag()
val computedVersionCode = maj * 10000 + min * 100 + pat
val computedVersionName = "$maj.$min.$pat"

plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
}

android {
    namespace = "ch.leys.bless2n"
    compileSdk = 36

    defaultConfig {
        applicationId = "ch.leys.bless2n"
        minSdk = 26
        targetSdk = 35
        versionName = computedVersionName
        versionCode = computedVersionCode

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        vectorDrawables { useSupportLibrary = true }

        val sumupAffiliateKey = (project.findProperty("sumupAffiliateKey") as String?)
            ?: System.getenv("SUMUP_AFFILIATE_KEY")
            ?: ""
        if (gradle.startParameter.taskNames.any {
                it.contains(
                    "Release",
                    ignoreCase = true
                )
            } && sumupAffiliateKey.isBlank()) {
            throw GradleException("SUMUP_AFFILIATE_KEY (or -PsumupAffiliateKey) must be set for release builds")
        }
        buildConfigField("String", "SUMUP_AFFILIATE_KEY", "\"$sumupAffiliateKey\"")

        buildConfigField("String", "POS_URL", "\"http://127.0.0.1:3000/pos\"")
        buildConfigField("boolean", "DEV_BUILD", "true")
    }

    buildFeatures {
        buildConfig = true
    }

    // Configure release signing via environment variables (set by CI)
    signingConfigs {
        create("release") {
            val buildingRelease =
                gradle.startParameter.taskNames.any { it.contains("Release", ignoreCase = true) }

            val storeFilePath = (System.getenv("BFS_UPLOAD_STORE_FILE")
                ?: project.findProperty("BFS_UPLOAD_STORE_FILE") as String?) ?: ""
            val storePasswordVal = (System.getenv("BFS_UPLOAD_STORE_PASSWORD")
                ?: project.findProperty("BFS_UPLOAD_STORE_PASSWORD") as String?) ?: ""
            val keyAliasVal = (System.getenv("BFS_UPLOAD_KEY_ALIAS")
                ?: project.findProperty("BFS_UPLOAD_KEY_ALIAS") as String?) ?: ""
            val keyPasswordVal = (System.getenv("BFS_UPLOAD_KEY_PASSWORD")
                ?: project.findProperty("BFS_UPLOAD_KEY_PASSWORD") as String?) ?: ""

            val missing =
                storeFilePath.isBlank() || storePasswordVal.isBlank() || keyAliasVal.isBlank() || keyPasswordVal.isBlank()
            if (buildingRelease && missing) {
                throw GradleException("Release signing missing. Set BFS_UPLOAD_STORE_FILE, BFS_UPLOAD_STORE_PASSWORD, BFS_UPLOAD_KEY_ALIAS, BFS_UPLOAD_KEY_PASSWORD.")
            }

            if (storeFilePath.isNotBlank()) storeFile = file(storeFilePath)
            if (storePasswordVal.isNotBlank()) storePassword = storePasswordVal
            if (keyAliasVal.isNotBlank()) keyAlias = keyAliasVal
            if (keyPasswordVal.isNotBlank()) keyPassword = keyPasswordVal
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )

            signingConfig = signingConfigs.getByName("release")

            val buildingRelease =
                gradle.startParameter.taskNames.any { it.contains("Release", ignoreCase = true) }
            val posUrl = (project.findProperty("posUrl") as String?) ?: System.getenv("POS_URL")
            if (buildingRelease && (posUrl.isNullOrBlank())) {
                throw GradleException("POS_URL (or -PposUrl) must be set for release builds")
            }
            val posUrlRelease = posUrl ?: "https://example.com/pos"
            buildConfigField("String", "POS_URL", "\"$posUrlRelease\"")
            buildConfigField("boolean", "DEV_BUILD", "false")

            // App name for Play/production builds
            resValue("string", "app_name", "BlessThun Food")
        }

        create("dev") {
            initWith(getByName("debug"))
            isDebuggable = true
            signingConfig = signingConfigs.getByName("debug")
            applicationIdSuffix = ".dev"
            versionNameSuffix = "-dev"

            val posUrl = (project.findProperty("posUrl") as String?)
                ?: System.getenv("POS_URL")
                ?: "http://127.0.0.1:3000/pos"
            buildConfigField("String", "POS_URL", "\"$posUrl\"")
            buildConfigField("boolean", "DEV_BUILD", "true")

            // App name for developer builds
            resValue("string", "app_name", "BlessThun Food (Dev)")
        }

        // Optional staging buildType for pre-prod testing
        create("staging") {
            initWith(getByName("release"))
            // Use debug signing by default for convenience; switch if needed
            signingConfig = signingConfigs.getByName("debug")
            applicationIdSuffix = ".staging"
            versionNameSuffix = "-staging"

            val posUrl = (project.findProperty("posUrl") as String?)
                ?: System.getenv("POS_URL")
                ?: "http://127.0.0.1:3000/pos"
            buildConfigField("String", "POS_URL", "\"$posUrl\"")
            buildConfigField("boolean", "DEV_BUILD", "false")

            // App name for staging builds
            resValue("string", "app_name", "BlessThun Food (Staging)")
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
        isCoreLibraryDesugaringEnabled = true
    }
    // Kotlin options moved to compilerOptions DSL below
    buildFeatures {
        compose = true
    }

    lint {
        // Vendor library (android-state) ships a lint registry without vendor metadata.
        // Do not fail release builds on such third-party lint issues.
        abortOnError = false
        checkReleaseBuilds = false
    }
}

// Kotlin compiler options (KGP 2.x)
kotlin {
    compilerOptions {
        jvmTarget.set(JvmTarget.JVM_11)
    }
}

dependencies {
    implementation("com.sumup:merchant-sdk:6.0.0")
    coreLibraryDesugaring("com.android.tools:desugar_jdk_libs:2.1.5")
    implementation("com.github.DantSu:ESCPOS-ThermalPrinter-Android:3.4.0")
    implementation("com.google.zxing:core:3.5.4")
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

    // Compile-only annotations used by transitive libraries to satisfy R8 analysis
    compileOnly("com.google.errorprone:error_prone_annotations:2.15.0")
}
