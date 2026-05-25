package com.yummi.app.ui.screens.recipes

import androidx.compose.animation.*
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.staggeredgrid.LazyVerticalStaggeredGrid
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridCells
import androidx.compose.foundation.lazy.staggeredgrid.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Clear
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.*
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.yummi.app.R
import com.yummi.app.data.api.ApiRecipe
import com.yummi.app.data.api.YummiApi
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RecipeListScreen(
    api: YummiApi,
    serverUrl: String,
    categorySlug: String? = null,
    tagSlug: String? = null,
    onRecipeClick: (Long) -> Unit,
    onImportClick: () -> Unit,
) {
    var recipes by remember { mutableStateOf<List<ApiRecipe>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }
    var isRefreshing by remember { mutableStateOf(false) }
    var searchQuery by remember { mutableStateOf("") }
    var searchJob by remember { mutableStateOf<Job?>(null) }
    val scope = rememberCoroutineScope()
    val snackbarHostState = remember { SnackbarHostState() }
    val connectionErrorMsg = stringResource(R.string.connection_error)

    suspend fun fetchRecipes(query: String? = null): List<ApiRecipe> {
        val response = if (!query.isNullOrBlank()) {
            api.listRecipes(query = query)
        } else if (!categorySlug.isNullOrBlank()) {
            api.listRecipes(category = categorySlug)
        } else if (!tagSlug.isNullOrBlank()) {
            api.listRecipes(tag = tagSlug)
        } else {
            api.listRecipes()
        }
        return if (response.isSuccessful) response.body() ?: emptyList() else emptyList()
    }

    // Initial load
    LaunchedEffect(Unit) {
        isLoading = true
        try {
            recipes = fetchRecipes()
        } catch (e: Exception) {
            snackbarHostState.showSnackbar(connectionErrorMsg)
        }
        isLoading = false
    }

    Box(modifier = Modifier.fillMaxSize()) {
        Column(modifier = Modifier.fillMaxSize()) {
            // Search bar
            SearchBar(
                query = searchQuery,
                onQueryChange = { query ->
                    searchQuery = query
                    searchJob?.cancel()
                    searchJob = scope.launch {
                        delay(300)
                        try {
                            recipes = fetchRecipes(query.takeIf { it.isNotBlank() })
                        } catch (_: Exception) {
                            recipes = emptyList()
                        }
                    }
                },
                onClear = {
                    searchQuery = ""
                    searchJob?.cancel()
                    scope.launch {
                        try { recipes = fetchRecipes() } catch (_: Exception) {}
                    }
                },
            )

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
                            try {
                                recipes = fetchRecipes(searchQuery.takeIf { it.isNotBlank() })
                            } catch (_: Exception) {}
                            isRefreshing = false
                        }
                    },
                    modifier = Modifier.fillMaxSize(),
                ) {
                    if (recipes.isEmpty()) {
                        Box(
                            modifier = Modifier.fillMaxSize(),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text(
                                text = if (searchQuery.isNotBlank()) stringResource(R.string.no_recipes_found) else stringResource(R.string.no_recipes_yet),
                                style = MaterialTheme.typography.bodyLarge,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                            )
                        }
                    } else {
                        LazyVerticalStaggeredGrid(
                            columns = StaggeredGridCells.Adaptive(minSize = 300.dp),
                            contentPadding = PaddingValues(16.dp),
                            verticalItemSpacing = 12.dp,
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            items(recipes, key = { it.id }) { recipe ->
                                com.yummi.app.ui.components.RecipeCard(
                                    recipe = recipe,
                                    serverUrl = serverUrl,
                                    onClick = { onRecipeClick(recipe.id) },
                                    modifier = Modifier.animateItem(),
                                )
                            }
                        }
                    }
                }
            }
        }

        FloatingActionButton(
            onClick = onImportClick,
            containerColor = MaterialTheme.colorScheme.primary,
            contentColor = MaterialTheme.colorScheme.onPrimary,
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .padding(16.dp),
        ) {
            Icon(Icons.Default.Add, contentDescription = stringResource(R.string.new_recipe))
        }

        SnackbarHost(
            hostState = snackbarHostState,
            modifier = Modifier.align(Alignment.BottomCenter),
        )
    }
}

@Composable
private fun SearchBar(
    query: String,
    onQueryChange: (String) -> Unit,
    onClear: () -> Unit,
) {
    OutlinedTextField(
        value = query,
        onValueChange = onQueryChange,
        placeholder = { Text(stringResource(R.string.search_recipes)) },
        leadingIcon = { Icon(Icons.Default.Search, contentDescription = null) },
        trailingIcon = {
            AnimatedVisibility(
                visible = query.isNotBlank(),
                enter = fadeIn(),
                exit = fadeOut(),
            ) {
                IconButton(onClick = onClear) {
                    Icon(Icons.Default.Clear, contentDescription = stringResource(R.string.clear_search))
                }
            }
        },
        singleLine = true,
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        shape = RoundedCornerShape(12.dp),
    )
}
