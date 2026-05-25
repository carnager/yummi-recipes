package com.yummi.app.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage
import com.yummi.app.data.api.ApiRecipe
import com.yummi.app.ui.theme.LightPlaceholder
import com.yummi.app.ui.theme.YummiOrangeLight

@Composable
fun RecipeCard(
    recipe: ApiRecipe,
    serverUrl: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val shape = RoundedCornerShape(12.dp)

    androidx.compose.material3.Card(
        modifier = modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
        shape = shape,
        colors = androidx.compose.material3.CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface,
        ),
        elevation = androidx.compose.material3.CardDefaults.cardElevation(defaultElevation = 2.dp),
    ) {
        Column {
            // Image
            if (recipe.imagePath.isNotBlank()) {
                AsyncImage(
                    model = "${serverUrl.trimEnd('/')}/uploads/${recipe.imagePath}",
                    contentDescription = recipe.title,
                    contentScale = ContentScale.Crop,
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(180.dp),
                )
            } else {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(180.dp)
                        .background(
                            Brush.linearGradient(
                                colors = listOf(
                                    MaterialTheme.colorScheme.surfaceVariant,
                                    LightPlaceholder,
                                )
                            )
                        ),
                    contentAlignment = Alignment.Center,
                ) {
                    Text("🍴", style = MaterialTheme.typography.headlineLarge)
                }
            }

            // Body
            Column(modifier = Modifier.padding(12.dp)) {
                Text(
                    text = recipe.title,
                    style = MaterialTheme.typography.titleLarge,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )

                if (recipe.description.isNotBlank()) {
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = recipe.description,
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis,
                    )
                }

                // Meta row
                val meta = buildList {
                    recipe.category?.let { add(it.name) }
                    if (recipe.prepTime.isNotBlank()) add("⏱ ${recipe.prepTime}")
                    if (recipe.servings.isNotBlank()) add("👥 ${recipe.servings}")
                }
                if (meta.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = meta.joinToString(" · "),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }

                // Tags
                if (!recipe.tags.isNullOrEmpty()) {
                    Spacer(modifier = Modifier.height(8.dp))
                    Row(
                        horizontalArrangement = Arrangement.spacedBy(6.dp),
                        modifier = Modifier.fillMaxWidth(),
                    ) {
                        recipe.tags.take(3).forEach { tag ->
                            TagChip(name = tag.name)
                        }
                        if (recipe.tags.size > 3) {
                            TagChip(name = "+${recipe.tags.size - 3}")
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun TagChip(name: String, modifier: Modifier = Modifier, onClick: (() -> Unit)? = null) {
    val shape = RoundedCornerShape(20.dp)
    val mod = if (onClick != null) modifier.clickable(onClick = onClick) else modifier

    Text(
        text = name,
        modifier = mod
            .background(YummiOrangeLight, shape)
            .padding(horizontal = 10.dp, vertical = 4.dp),
        style = MaterialTheme.typography.labelSmall,
        color = MaterialTheme.colorScheme.primary,
    )
}
