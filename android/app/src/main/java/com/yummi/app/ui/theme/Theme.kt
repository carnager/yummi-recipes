package com.yummi.app.ui.theme

import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

private val LightColorScheme = lightColorScheme(
    primary = YummiOrange,
    onPrimary = YummiWhite,
    primaryContainer = YummiOrangeLight,
    onPrimaryContainer = YummiOrange,
    secondary = YummiOrange,
    onSecondary = YummiWhite,
    background = LightBackground,
    onBackground = LightText,
    surface = LightSurface,
    onSurface = LightText,
    surfaceVariant = LightBackground,
    onSurfaceVariant = LightTextMuted,
    surfaceContainerLowest = LightSurface,
    surfaceContainerLow = LightSurface,
    surfaceContainer = LightBackground,
    surfaceContainerHigh = LightBackground,
    surfaceContainerHighest = LightBorder,
    outline = LightBorder,
    outlineVariant = LightBorder,
    error = YummiDanger,
    onError = YummiWhite,
    tertiary = YummiSuccess,
    onTertiary = YummiWhite,
)

private val DarkColorScheme = darkColorScheme(
    primary = YummiOrange,
    onPrimary = YummiWhite,
    primaryContainer = YummiOrangeMedium,
    onPrimaryContainer = YummiOrange,
    secondary = YummiOrange,
    onSecondary = YummiWhite,
    background = DarkBackground,
    onBackground = DarkText,
    surface = DarkSurface,
    onSurface = DarkText,
    surfaceVariant = DarkBackground,
    onSurfaceVariant = DarkTextMuted,
    outline = DarkBorder,
    outlineVariant = DarkBorder,
    error = YummiDanger,
    onError = YummiWhite,
    tertiary = YummiSuccess,
    onTertiary = YummiWhite,
    surfaceContainer = DarkNavBg,
    surfaceContainerHigh = DarkSurface,
)

@Composable
fun YummiTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit,
) {
    val colorScheme = if (darkTheme) DarkColorScheme else LightColorScheme

    MaterialTheme(
        colorScheme = colorScheme,
        typography = YummiTypography,
        content = content,
    )
}
