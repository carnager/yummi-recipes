package com.yummi.app.data.api

import com.yummi.app.data.local.PreferencesManager
import kotlinx.coroutines.flow.firstOrNull
import kotlinx.coroutines.runBlocking
import okhttp3.Interceptor
import okhttp3.Response

class AuthInterceptor(private val prefs: PreferencesManager) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val original = chain.request()

        // Skip auth for login/register
        if (original.url.encodedPath.contains("/auth/")) {
            return chain.proceed(original)
        }

        val token = runBlocking { prefs.tokenFlow.firstOrNull() }
        if (token.isNullOrBlank()) {
            return chain.proceed(original)
        }

        val request = original.newBuilder()
            .header("Authorization", "Bearer $token")
            .build()
        return chain.proceed(request)
    }
}
