pluginManagement {
    repositories {
        google {
            content {
                includeGroupByRegex("com\\.android.*")
                includeGroupByRegex("com\\.google.*")
                includeGroupByRegex("androidx.*")
            }
        }
        mavenCentral()
        gradlePluginPortal()
    }
}
dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
        // SumUp private Maven
        maven { url = uri("https://maven.sumup.com/releases") }
        // JitPack for DantSu
        maven { url = uri("https://jitpack.io") }
    }
}

rootProject.name = "bfs-android-app"
include(":app")
 