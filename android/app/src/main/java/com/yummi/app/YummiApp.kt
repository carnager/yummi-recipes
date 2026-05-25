package com.yummi.app

import android.app.Application
import com.yummi.app.data.api.AuthInterceptor
import com.yummi.app.data.api.YummiApi
import com.yummi.app.data.local.PreferencesManager
import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import kotlinx.coroutines.flow.firstOrNull
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import java.util.concurrent.TimeUnit

class YummiApp : Application() {

    lateinit var prefs: PreferencesManager
        private set

    private var _api: YummiApi? = null
    private var _currentBaseUrl: String? = null

    val json = Json {
        ignoreUnknownKeys = true
        coerceInputValues = true
        isLenient = true
    }

    fun getApi(): YummiApi? = _api

    fun buildApi(baseUrl: String): YummiApi {
        val url = baseUrl.trimEnd('/') + "/"
        if (_api != null && _currentBaseUrl == url) return _api!!

        val client = OkHttpClient.Builder()
            .addInterceptor(AuthInterceptor(prefs))
            .addInterceptor(HttpLoggingInterceptor().apply {
                level = HttpLoggingInterceptor.Level.BODY
            })
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(60, TimeUnit.SECONDS)
            .build()

        val retrofit = Retrofit.Builder()
            .baseUrl(url)
            .client(client)
            .addConverterFactory(json.asConverterFactory("application/json".toMediaType()))
            .build()

        _api = retrofit.create(YummiApi::class.java)
        _currentBaseUrl = url
        return _api!!
    }

    override fun onCreate() {
        super.onCreate()
        prefs = PreferencesManager(this)

        // Restore API from saved server URL
        val savedUrl = runBlocking { prefs.serverUrlFlow.firstOrNull() }
        if (!savedUrl.isNullOrBlank()) {
            buildApi(savedUrl)
        }
    }
}
