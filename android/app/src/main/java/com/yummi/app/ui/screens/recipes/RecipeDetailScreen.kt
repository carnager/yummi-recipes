package com.yummi.app.ui.screens.recipes

import android.view.HapticFeedbackConstants
import androidx.compose.animation.*
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalView
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage
import com.yummi.app.data.api.ApiRecipe
import com.yummi.app.data.api.YummiApi
import com.yummi.app.ui.components.MarkdownText
import com.yummi.app.ui.components.ShareBottomSheet
import com.yummi.app.ui.components.TagChip
import com.yummi.app.ui.theme.YummiSuccess
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun RecipeDetailScreen(
    recipeId: Long,
    api: YummiApi,
    serverUrl: String,
    currentUserId: Long,
    onEdit: (Long) -> Unit,
    onDeleted: () -> Unit,
    onCategoryClick: (String) -> Unit,
    onTagClick: (String) -> Unit,
) {
    var recipe by remember { mutableStateOf<ApiRecipe?>(null) }
    var isLoading by remember { mutableStateOf(true) }
    var tried by remember { mutableStateOf(false) }
    var showDeleteDialog by remember { mutableStateOf(false) }
    var showShareSheet by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()
    val snackbarHostState = remember { SnackbarHostState() }
    val view = LocalView.current

    LaunchedEffect(recipeId) {
        isLoading = true
        try {
            val response = api.getRecipe(recipeId)
            if (response.isSuccessful) {
                recipe = response.body()
                tried = recipe?.tried ?: false
            }
        } catch (_: Exception) {}
        isLoading = false
    }

    if (showShareSheet && recipe != null) {
        ShareBottomSheet(
            recipeId = recipe!!.id,
            currentUserId = currentUserId,
            api = api,
            onDismiss = { showShareSheet = false },
        )
    }

    if (showDeleteDialog) {
        AlertDialog(
            onDismissRequest = { showDeleteDialog = false },
            icon = { Icon(Icons.Default.DeleteForever, contentDescription = null, tint = MaterialTheme.colorScheme.error) },
            title = { Text("Rezept loeschen?") },
            text = { Text("Dieses Rezept wird unwiderruflich geloescht.") },
            confirmButton = {
                TextButton(
                    onClick = {
                        scope.launch {
                            try {
                                api.deleteRecipe(recipeId)
                                onDeleted()
                            } catch (_: Exception) {
                                snackbarHostState.showSnackbar("Loeschen fehlgeschlagen")
                            }
                        }
                    },
                    colors = ButtonDefaults.textButtonColors(contentColor = MaterialTheme.colorScheme.error),
                ) { Text("Loeschen") }
            },
            dismissButton = {
                TextButton(onClick = { showDeleteDialog = false }) { Text("Abbrechen") }
            },
        )
    }

    Box(modifier = Modifier.fillMaxSize()) {
        SnackbarHost(
            hostState = snackbarHostState,
            modifier = Modifier.align(Alignment.BottomCenter),
        )
        if (isLoading) {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center,
            ) {
                CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
            }
        } else if (recipe == null) {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center,
            ) {
                Text("Rezept nicht gefunden", style = MaterialTheme.typography.bodyLarge)
            }
        } else {
            val r = recipe!!
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState()),
            ) {
                // Hero image with rounded bottom corners
                if (r.imagePath.isNotBlank()) {
                    AsyncImage(
                        model = "${serverUrl.trimEnd('/')}/uploads/${r.imagePath}",
                        contentDescription = r.title,
                        contentScale = ContentScale.Crop,
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(280.dp)
                            .clip(RoundedCornerShape(bottomStart = 24.dp, bottomEnd = 24.dp)),
                    )
                }

                Column(modifier = Modifier.padding(16.dp)) {
                    // Title
                    Text(
                        text = r.title,
                        style = MaterialTheme.typography.headlineMedium,
                    )

                    if (r.description.isNotBlank()) {
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = r.description,
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }

                    // Meta chips
                    Spacer(modifier = Modifier.height(16.dp))
                    FlowRow(
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        r.category?.let { cat ->
                            AssistChip(
                                onClick = { onCategoryClick(cat.slug) },
                                label = { Text(cat.name) },
                                leadingIcon = { Text("📂", style = MaterialTheme.typography.bodySmall) },
                                shape = RoundedCornerShape(20.dp),
                            )
                        }
                        if (r.prepTime.isNotBlank()) {
                            AssistChip(
                                onClick = {},
                                label = { Text(r.prepTime) },
                                leadingIcon = { Text("⏱", style = MaterialTheme.typography.bodySmall) },
                                shape = RoundedCornerShape(20.dp),
                            )
                        }
                        if (r.cookTime.isNotBlank()) {
                            AssistChip(
                                onClick = {},
                                label = { Text(r.cookTime) },
                                leadingIcon = { Text("🔥", style = MaterialTheme.typography.bodySmall) },
                                shape = RoundedCornerShape(20.dp),
                            )
                        }
                        if (r.servings.isNotBlank()) {
                            AssistChip(
                                onClick = {},
                                label = { Text(r.servings) },
                                leadingIcon = { Text("👥", style = MaterialTheme.typography.bodySmall) },
                                shape = RoundedCornerShape(20.dp),
                            )
                        }
                    }

                    // Tags
                    if (!r.tags.isNullOrEmpty()) {
                        Spacer(modifier = Modifier.height(12.dp))
                        FlowRow(
                            horizontalArrangement = Arrangement.spacedBy(6.dp),
                            verticalArrangement = Arrangement.spacedBy(6.dp),
                        ) {
                            r.tags.forEach { tag ->
                                TagChip(name = tag.name, onClick = { onTagClick(tag.slug) })
                            }
                        }
                    }

                    // Action buttons
                    Spacer(modifier = Modifier.height(20.dp))
                    FlowRow(
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        // Tried button
                        FilledTonalButton(
                            onClick = {
                                view.performHapticFeedback(HapticFeedbackConstants.CONFIRM)
                                scope.launch {
                                    try {
                                        val resp = api.toggleTried(r.id)
                                        if (resp.isSuccessful) {
                                            tried = resp.body()?.tried ?: !tried
                                        }
                                    } catch (_: Exception) {}
                                }
                            },
                            colors = if (tried) {
                                ButtonDefaults.filledTonalButtonColors(
                                    containerColor = YummiSuccess,
                                    contentColor = MaterialTheme.colorScheme.onPrimary,
                                )
                            } else {
                                ButtonDefaults.filledTonalButtonColors()
                            },
                            shape = RoundedCornerShape(12.dp),
                        ) {
                            AnimatedContent(targetState = tried, label = "tried") { isTried ->
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                                ) {
                                    Icon(
                                        if (isTried) Icons.Default.Check else Icons.Default.Restaurant,
                                        contentDescription = null,
                                        modifier = Modifier.size(18.dp),
                                    )
                                    Text(if (isTried) "Probiert!" else "Probiert?")
                                }
                            }
                        }

                        // Share button
                        if (r.createdBy == currentUserId) {
                            FilledTonalButton(
                                onClick = {
                                    view.performHapticFeedback(HapticFeedbackConstants.CONTEXT_CLICK)
                                    showShareSheet = true
                                },
                                shape = RoundedCornerShape(12.dp),
                            ) {
                                Icon(
                                    Icons.Default.Share,
                                    contentDescription = null,
                                    modifier = Modifier.size(18.dp),
                                )
                                Spacer(modifier = Modifier.width(6.dp))
                                Text("Teilen")
                            }

                            FilledTonalButton(
                                onClick = { onEdit(r.id) },
                                shape = RoundedCornerShape(12.dp),
                            ) {
                                Icon(
                                    Icons.Default.Edit,
                                    contentDescription = null,
                                    modifier = Modifier.size(18.dp),
                                )
                                Spacer(modifier = Modifier.width(6.dp))
                                Text("Bearbeiten")
                            }

                            FilledTonalButton(
                                onClick = { showDeleteDialog = true },
                                colors = ButtonDefaults.filledTonalButtonColors(
                                    containerColor = MaterialTheme.colorScheme.errorContainer,
                                    contentColor = MaterialTheme.colorScheme.error,
                                ),
                                shape = RoundedCornerShape(12.dp),
                            ) {
                                Icon(
                                    Icons.Default.Delete,
                                    contentDescription = null,
                                    modifier = Modifier.size(18.dp),
                                )
                                Spacer(modifier = Modifier.width(6.dp))
                                Text("Löschen")
                            }
                        }
                    }

                    // Divider
                    Spacer(modifier = Modifier.height(20.dp))
                    HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant)
                    Spacer(modifier = Modifier.height(20.dp))

                    // Source URL
                    if (r.sourceUrl.isNotBlank()) {
                        Surface(
                            shape = RoundedCornerShape(12.dp),
                            color = MaterialTheme.colorScheme.surfaceVariant,
                            modifier = Modifier.fillMaxWidth(),
                        ) {
                            Row(
                                modifier = Modifier.padding(12.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(8.dp),
                            ) {
                                Icon(
                                    Icons.Default.Link,
                                    contentDescription = null,
                                    tint = MaterialTheme.colorScheme.primary,
                                    modifier = Modifier.size(18.dp),
                                )
                                Text(
                                    text = r.sourceUrl,
                                    style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.primary,
                                )
                            }
                        }
                        Spacer(modifier = Modifier.height(20.dp))
                    }

                    // Content
                    if (r.contentMd.isNotBlank()) {
                        MarkdownText(
                            markdown = r.contentMd,
                            modifier = Modifier.fillMaxWidth(),
                        )
                    }

                    Spacer(modifier = Modifier.height(32.dp))
                }
            }
        }
    }
}
