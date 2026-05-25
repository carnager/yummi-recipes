package com.yummi.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.runtime.*
import androidx.compose.ui.platform.LocalContext
import com.yummi.app.data.api.LoginRequest
import com.yummi.app.data.api.RegisterRequest
import com.yummi.app.ui.navigation.YummiNavGraph
import com.yummi.app.ui.screens.login.LoginScreen
import com.yummi.app.ui.theme.YummiTheme
import kotlinx.coroutines.flow.firstOrNull
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()

        setContent {
            YummiRoot()
        }
    }
}

@Composable
fun YummiRoot() {
    val context = LocalContext.current
    val app = context.applicationContext as YummiApp
    val systemDark = isSystemInDarkTheme()

    var isLoggedIn by remember { mutableStateOf<Boolean?>(null) }
    var currentUserId by remember { mutableStateOf(0L) }
    var serverUrl by remember { mutableStateOf("") }
    var initialServerUrl by remember { mutableStateOf("") }
    var isDarkMode by remember { mutableStateOf(systemDark) }
    val scope = rememberCoroutineScope()

    // Load saved preferences
    LaunchedEffect(Unit) {
        val token = app.prefs.tokenFlow.firstOrNull()
        val savedUrl = app.prefs.serverUrlFlow.firstOrNull()
        val savedDark = app.prefs.darkModeFlow.firstOrNull()
        currentUserId = app.prefs.userIdFlow.firstOrNull() ?: 0
        initialServerUrl = savedUrl ?: ""
        serverUrl = savedUrl ?: ""
        isDarkMode = savedDark ?: systemDark

        isLoggedIn = !token.isNullOrBlank() && !savedUrl.isNullOrBlank() && app.getApi() != null
    }

    YummiTheme(darkTheme = isDarkMode) {
        when (isLoggedIn) {
            null -> {
                // Loading
            }
            false -> {
                LoginScreen(
                    initialServerUrl = initialServerUrl,
                    isDarkMode = isDarkMode,
                    onToggleTheme = {
                        isDarkMode = !isDarkMode
                        scope.launch { app.prefs.setDarkMode(isDarkMode) }
                    },
                    onLogin = { url, username, password ->
                        try {
                            val api = app.buildApi(url)
                            val resp = api.login(LoginRequest(username, password))
                            if (resp.isSuccessful) {
                                val body = resp.body()!!
                                app.prefs.saveServerUrl(url)
                                app.prefs.saveAuth(body.token, body.user.id, body.user.username, body.user.displayName)
                                serverUrl = url
                                currentUserId = body.user.id
                                isLoggedIn = true
                                Result.success(Unit)
                            } else {
                                Result.failure(Exception("Benutzername oder Passwort falsch"))
                            }
                        } catch (e: Exception) {
                            Result.failure(Exception("Verbindung fehlgeschlagen: ${e.message}"))
                        }
                    },
                    onRegister = { url, username, displayName, password ->
                        try {
                            val api = app.buildApi(url)
                            val resp = api.register(RegisterRequest(username, displayName, password))
                            if (resp.isSuccessful) {
                                val body = resp.body()!!
                                app.prefs.saveServerUrl(url)
                                app.prefs.saveAuth(body.token, body.user.id, body.user.username, body.user.displayName)
                                serverUrl = url
                                currentUserId = body.user.id
                                isLoggedIn = true
                                Result.success(Unit)
                            } else {
                                Result.failure(Exception("Registrierung fehlgeschlagen"))
                            }
                        } catch (e: Exception) {
                            Result.failure(Exception("Verbindung fehlgeschlagen: ${e.message}"))
                        }
                    },
                )
            }
            true -> {
                val api = app.getApi()
                if (api != null) {
                    YummiNavGraph(
                        api = api,
                        serverUrl = serverUrl,
                        currentUserId = currentUserId,
                        isDarkMode = isDarkMode,
                        onToggleTheme = {
                            isDarkMode = !isDarkMode
                            scope.launch { app.prefs.setDarkMode(isDarkMode) }
                        },
                        onLogout = {
                            scope.launch {
                                app.prefs.clear()
                                isLoggedIn = false
                            }
                        },
                    )
                }
            }
        }
    }
}
