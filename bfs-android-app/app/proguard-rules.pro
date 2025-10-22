# Add project specific ProGuard rules here.
# You can control the set of applied configuration files using the
# proguardFiles setting in build.gradle.
#
# For more details, see
#   http://developer.android.com/guide/developing/tools/proguard.html

# If your project uses WebView with JS, uncomment the following
# and specify the fully qualified class name to the JavaScript interface
# class:
#-keepclassmembers class fqcn.of.javascript.interface.for.webview {
#   public *;
#}

# Uncomment this to preserve the line number information for
# debugging stack traces.
#-keepattributes SourceFile,LineNumberTable

# If you keep the line number information, uncomment this to
# hide the original source file name.
#-renamesourcefileattribute SourceFile

# --- Suppress missing/optional dependencies referenced by vendor SDKs ---
# SumUp SDK references optional analytics/observability and other libs that are not bundled.
# We don't use these features, so suppress warnings about their absence during shrink.
-dontwarn com.google.errorprone.annotations.**
-dontwarn com.sumup.analyticskit.**
-dontwarn com.sumup.mixpanel.**
-dontwarn com.sumup.observabilitylib.**
-dontwarn com.sumup.observablib.**
-dontwarn io.opentelemetry.**
-dontwarn org.greenrobot.eventbus.**
-dontwarn com.google.firebase.**
-dontwarn com.mixpanel.**
-dontwarn kotlinx.parcelize.**

# Keep annotations and signatures to avoid issues with Kotlin/Parcelize reflection
-keepattributes *Annotation*,Signature
