package com.yummi.app.ui.screens.categories

import androidx.compose.animation.animateContentSize
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.yummi.app.data.api.ApiCategory
import com.yummi.app.data.api.YummiApi
import kotlinx.coroutines.launch

private val categoryIcons = mapOf(
    "vorspeisen" to "🥗", "suppen" to "🍜", "salate" to "🥬",
    "hauptgerichte" to "🍽️", "beilagen" to "🥔", "desserts" to "🍰",
    "backen" to "🧁", "getraenke" to "🥤", "fruehstueck" to "🍳",
    "snacks" to "🥨", "saucen" to "🫙", "eingemachtes" to "🏺",
    "grillen" to "🔥", "vegetarisch" to "🥦", "vegan" to "🌱",
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CategoriesScreen(
    api: YummiApi,
    onCategoryClick: (String) -> Unit,
) {
    var categories by remember { mutableStateOf<List<ApiCategory>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }
    var isRefreshing by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()

    suspend fun load() {
        try {
            val resp = api.listCategories()
            if (resp.isSuccessful) categories = resp.body() ?: emptyList()
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
            LazyVerticalGrid(
                columns = GridCells.Adaptive(minSize = 140.dp),
                contentPadding = PaddingValues(16.dp),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                items(categories, key = { it.id }) { cat ->
                    ElevatedCard(
                        modifier = Modifier
                            .fillMaxWidth()
                            .animateContentSize()
                            .clickable { onCategoryClick(cat.slug) },
                        shape = RoundedCornerShape(16.dp),
                    ) {
                        Column(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(16.dp),
                            horizontalAlignment = Alignment.CenterHorizontally,
                        ) {
                            Surface(
                                shape = RoundedCornerShape(12.dp),
                                color = MaterialTheme.colorScheme.primaryContainer,
                                modifier = Modifier.size(56.dp),
                            ) {
                                Box(contentAlignment = Alignment.Center, modifier = Modifier.fillMaxSize()) {
                                    Text(
                                        text = categoryIcons[cat.slug] ?: "📁",
                                        style = MaterialTheme.typography.headlineMedium,
                                    )
                                }
                            }
                            Spacer(modifier = Modifier.height(12.dp))
                            Text(
                                text = cat.name,
                                style = MaterialTheme.typography.titleMedium,
                                fontWeight = FontWeight.SemiBold,
                                textAlign = TextAlign.Center,
                            )
                        }
                    }
                }
            }
        }
    }
}
