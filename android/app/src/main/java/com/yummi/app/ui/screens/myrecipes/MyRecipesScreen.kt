package com.yummi.app.ui.screens.myrecipes

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.yummi.app.data.api.ApiRecipe
import com.yummi.app.data.api.YummiApi
import com.yummi.app.ui.components.RecipeCard
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MyRecipesScreen(
    api: YummiApi,
    serverUrl: String,
    onRecipeClick: (Long) -> Unit,
) {
    var ownRecipes by remember { mutableStateOf<List<ApiRecipe>>(emptyList()) }
    var sharedRecipes by remember { mutableStateOf<List<ApiRecipe>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }
    var isRefreshing by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()

    suspend fun load() {
        try {
            val resp = api.myRecipes()
            if (resp.isSuccessful) {
                val body = resp.body()
                ownRecipes = body?.own ?: emptyList()
                sharedRecipes = body?.shared ?: emptyList()
            }
        } catch (_: Exception) {}
    }

    LaunchedEffect(Unit) {
        load()
        isLoading = false
    }

    if (isLoading) {
        Box(
            modifier = Modifier.fillMaxSize(),
            contentAlignment = Alignment.Center,
        ) {
            CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
        }
    } else {
        PullToRefreshBox(
            isRefreshing = isRefreshing,
            onRefresh = {
                isRefreshing = true
                scope.launch {
                    load()
                    isRefreshing = false
                }
            },
            modifier = Modifier.fillMaxSize(),
        ) {
            if (ownRecipes.isEmpty() && sharedRecipes.isEmpty()) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        text = "Noch keine Rezepte",
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
            } else {
                LazyColumn(
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    if (ownRecipes.isNotEmpty()) {
                        item(key = "header-own") {
                            SectionHeader("Meine Rezepte", ownRecipes.size)
                        }
                        items(ownRecipes, key = { "own-${it.id}" }) { recipe ->
                            RecipeCard(
                                recipe = recipe,
                                serverUrl = serverUrl,
                                onClick = { onRecipeClick(recipe.id) },
                            )
                        }
                    }

                    if (sharedRecipes.isNotEmpty()) {
                        item(key = "header-shared") {
                            Spacer(modifier = Modifier.height(8.dp))
                            SectionHeader("Mit mir geteilt", sharedRecipes.size)
                        }
                        items(sharedRecipes, key = { "shared-${it.id}" }) { recipe ->
                            RecipeCard(
                                recipe = recipe,
                                serverUrl = serverUrl,
                                onClick = { onRecipeClick(recipe.id) },
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun SectionHeader(title: String, count: Int) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(bottom = 4.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            text = title,
            style = MaterialTheme.typography.titleLarge,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        Badge(containerColor = MaterialTheme.colorScheme.primaryContainer) {
            Text(
                text = count.toString(),
                color = MaterialTheme.colorScheme.onPrimaryContainer,
            )
        }
    }
    HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant)
}
