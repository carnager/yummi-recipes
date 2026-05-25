package com.yummi.app.ui.screens.recipes

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Save
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.yummi.app.R
import com.yummi.app.data.api.*
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RecipeFormScreen(
    recipeId: Long?,
    api: YummiApi,
    onBack: () -> Unit,
    onSaved: (Long) -> Unit,
) {
    var title by remember { mutableStateOf("") }
    var description by remember { mutableStateOf("") }
    var sourceUrl by remember { mutableStateOf("") }
    var prepTime by remember { mutableStateOf("") }
    var cookTime by remember { mutableStateOf("") }
    var servings by remember { mutableStateOf("") }
    var contentMd by remember { mutableStateOf("") }
    var tags by remember { mutableStateOf("") }
    var categories by remember { mutableStateOf<List<ApiCategory>>(emptyList()) }
    var selectedCategoryId by remember { mutableStateOf<Long?>(null) }
    var categoryExpanded by remember { mutableStateOf(false) }
    var isLoading by remember { mutableStateOf(recipeId != null) }
    var isSaving by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()
    val snackbarHostState = remember { SnackbarHostState() }
    val saveErrorMsg = stringResource(R.string.save_error)

    LaunchedEffect(Unit) {
        try {
            val catResp = api.listCategories()
            if (catResp.isSuccessful) categories = catResp.body() ?: emptyList()
        } catch (_: Exception) {}

        if (recipeId != null) {
            try {
                val resp = api.getRecipe(recipeId)
                if (resp.isSuccessful) {
                    val r = resp.body()!!
                    title = r.title
                    description = r.description
                    sourceUrl = r.sourceUrl
                    prepTime = r.prepTime
                    cookTime = r.cookTime
                    servings = r.servings
                    contentMd = r.contentMd
                    selectedCategoryId = r.categoryId
                    tags = r.tags?.joinToString(", ") { it.name } ?: ""
                }
            } catch (_: Exception) {}
        }
        isLoading = false
    }

    Box(modifier = Modifier.fillMaxSize()) {
        if (isLoading) {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center,
            ) {
                CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
            }
        } else {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState())
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                OutlinedTextField(
                    value = title,
                    onValueChange = { title = it },
                    label = { Text(stringResource(R.string.title)) },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    singleLine = true,
                )

                OutlinedTextField(
                    value = description,
                    onValueChange = { description = it },
                    label = { Text(stringResource(R.string.description)) },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    maxLines = 3,
                )

                // Category dropdown
                ExposedDropdownMenuBox(
                    expanded = categoryExpanded,
                    onExpandedChange = { categoryExpanded = !categoryExpanded },
                ) {
                    OutlinedTextField(
                        value = categories.find { it.id == selectedCategoryId }?.name ?: "",
                        onValueChange = {},
                        readOnly = true,
                        label = { Text(stringResource(R.string.category)) },
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = categoryExpanded) },
                        modifier = Modifier.fillMaxWidth().menuAnchor(),
                        shape = RoundedCornerShape(12.dp),
                    )
                    ExposedDropdownMenu(
                        expanded = categoryExpanded,
                        onDismissRequest = { categoryExpanded = false },
                    ) {
                        DropdownMenuItem(
                            text = { Text(stringResource(R.string.no_category)) },
                            onClick = {
                                selectedCategoryId = null
                                categoryExpanded = false
                            },
                        )
                        categories.forEach { cat ->
                            DropdownMenuItem(
                                text = { Text(cat.name) },
                                onClick = {
                                    selectedCategoryId = cat.id
                                    categoryExpanded = false
                                },
                            )
                        }
                    }
                }

                // Times & servings
                Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    OutlinedTextField(
                        value = prepTime,
                        onValueChange = { prepTime = it },
                        label = { Text(stringResource(R.string.prep_time)) },
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(12.dp),
                        singleLine = true,
                    )
                    OutlinedTextField(
                        value = cookTime,
                        onValueChange = { cookTime = it },
                        label = { Text(stringResource(R.string.cook_time)) },
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(12.dp),
                        singleLine = true,
                    )
                }

                OutlinedTextField(
                    value = servings,
                    onValueChange = { servings = it },
                    label = { Text(stringResource(R.string.servings)) },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    singleLine = true,
                )

                OutlinedTextField(
                    value = sourceUrl,
                    onValueChange = { sourceUrl = it },
                    label = { Text(stringResource(R.string.source_url)) },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    singleLine = true,
                )

                OutlinedTextField(
                    value = tags,
                    onValueChange = { tags = it },
                    label = { Text(stringResource(R.string.tags_comma)) },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    singleLine = true,
                )

                OutlinedTextField(
                    value = contentMd,
                    onValueChange = { contentMd = it },
                    label = { Text(stringResource(R.string.content_markdown)) },
                    modifier = Modifier
                        .fillMaxWidth()
                        .heightIn(min = 200.dp),
                    shape = RoundedCornerShape(12.dp),
                )

                // Space for FAB
                Spacer(modifier = Modifier.height(72.dp))
            }
        }

        // FAB
        ExtendedFloatingActionButton(
            onClick = {
                if (title.isBlank() || isSaving) return@ExtendedFloatingActionButton
                isSaving = true
                scope.launch {
                    try {
                        val req = CreateRecipeRequest(
                            title = title,
                            description = description,
                            sourceUrl = sourceUrl,
                            prepTime = prepTime,
                            cookTime = cookTime,
                            servings = servings,
                            contentMd = contentMd,
                            categoryId = selectedCategoryId,
                            tags = tags,
                        )
                        val resp = if (recipeId != null) {
                            api.updateRecipe(recipeId, req)
                        } else {
                            api.createRecipe(req)
                        }
                        if (resp.isSuccessful) {
                            val saved = resp.body()
                            onSaved(saved?.id ?: recipeId ?: 0)
                        } else {
                            snackbarHostState.showSnackbar(saveErrorMsg)
                        }
                    } catch (e: Exception) {
                        snackbarHostState.showSnackbar(e.message ?: saveErrorMsg)
                    }
                    isSaving = false
                }
            },
            containerColor = MaterialTheme.colorScheme.primary,
            contentColor = MaterialTheme.colorScheme.onPrimary,
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .padding(16.dp),
        ) {
            if (isSaving) {
                CircularProgressIndicator(
                    modifier = Modifier.size(18.dp),
                    strokeWidth = 2.dp,
                    color = MaterialTheme.colorScheme.onPrimary,
                )
            } else {
                Icon(Icons.Default.Save, contentDescription = null)
            }
            Spacer(modifier = Modifier.width(8.dp))
            Text(stringResource(R.string.save))
        }

        SnackbarHost(
            hostState = snackbarHostState,
            modifier = Modifier.align(Alignment.BottomCenter),
        )
    }
}
