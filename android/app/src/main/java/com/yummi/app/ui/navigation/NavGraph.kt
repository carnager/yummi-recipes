package com.yummi.app.ui.navigation

import androidx.compose.animation.*
import androidx.compose.animation.core.tween
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.Logout
import androidx.compose.material.icons.filled.*
import androidx.compose.material.icons.outlined.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.NavType
import androidx.navigation.compose.*
import androidx.navigation.navArgument
import com.yummi.app.data.api.YummiApi
import com.yummi.app.ui.screens.categories.CategoriesScreen
import com.yummi.app.ui.screens.import_recipe.ImportScreen
import com.yummi.app.ui.screens.myrecipes.MyRecipesScreen
import com.yummi.app.ui.screens.recipes.*

sealed class Screen(val route: String) {
    data object Recipes : Screen("recipes")
    data object Categories : Screen("categories")
    data object MyRecipes : Screen("my-recipes")
    data object RecipeDetail : Screen("recipes/{id}") {
        fun create(id: Long) = "recipes/$id"
    }
    data object RecipeCreate : Screen("recipes/create")
    data object RecipeEdit : Screen("recipes/{id}/edit") {
        fun create(id: Long) = "recipes/$id/edit"
    }
    data object CategoryRecipes : Screen("categories/{slug}") {
        fun create(slug: String) = "categories/$slug"
    }
    data object TagRecipes : Screen("tags/{slug}") {
        fun create(slug: String) = "tags/$slug"
    }
    data object Import : Screen("import")
}

data class BottomNavItem(
    val label: String,
    val selectedIcon: ImageVector,
    val unselectedIcon: ImageVector,
    val route: String,
)

val bottomNavItems = listOf(
    BottomNavItem("Rezepte", Icons.Filled.Restaurant, Icons.Outlined.Restaurant, Screen.Recipes.route),
    BottomNavItem("Kategorien", Icons.Filled.Category, Icons.Outlined.Category, Screen.Categories.route),
    BottomNavItem("Meine", Icons.Filled.Bookmark, Icons.Outlined.BookmarkBorder, Screen.MyRecipes.route),
)

private val enterTransition: EnterTransition = fadeIn(tween(200)) + slideInHorizontally(tween(250)) { it / 4 }
private val exitTransition: ExitTransition = fadeOut(tween(200))
private val popEnterTransition: EnterTransition = fadeIn(tween(200)) + slideInHorizontally(tween(250)) { -it / 4 }
private val popExitTransition: ExitTransition = fadeOut(tween(200)) + slideOutHorizontally(tween(250)) { it / 4 }

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun YummiNavGraph(
    api: YummiApi,
    serverUrl: String,
    currentUserId: Long,
    isDarkMode: Boolean,
    onToggleTheme: () -> Unit,
    onLogout: () -> Unit,
) {
    val navController = rememberNavController()
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    val showBottomBar = currentRoute in bottomNavItems.map { it.route }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        text = "Yummi",
                        style = MaterialTheme.typography.headlineSmall,
                        fontWeight = FontWeight.ExtraBold,
                        color = MaterialTheme.colorScheme.primary,
                    )
                },
                navigationIcon = {
                    if (!showBottomBar) {
                        IconButton(onClick = { navController.popBackStack() }) {
                            Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Zurück")
                        }
                    }
                },
                actions = {
                    IconButton(onClick = onToggleTheme) {
                        Icon(
                            if (isDarkMode) Icons.Filled.LightMode else Icons.Filled.DarkMode,
                            contentDescription = "Theme wechseln",
                        )
                    }
                    if (showBottomBar) {
                        IconButton(onClick = onLogout) {
                            Icon(Icons.AutoMirrored.Filled.Logout, contentDescription = "Abmelden")
                        }
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.surface,
                ),
            )
        },
        bottomBar = {
            if (showBottomBar) {
                NavigationBar(
                    containerColor = MaterialTheme.colorScheme.surface,
                    tonalElevation = 0.dp,
                ) {
                    bottomNavItems.forEach { item ->
                        val selected = currentRoute == item.route
                        NavigationBarItem(
                            selected = selected,
                            onClick = {
                                navController.navigate(item.route) {
                                    popUpTo(navController.graph.findStartDestination().id) {
                                        saveState = true
                                    }
                                    launchSingleTop = true
                                    restoreState = true
                                }
                            },
                            icon = {
                                Icon(
                                    if (selected) item.selectedIcon else item.unselectedIcon,
                                    contentDescription = item.label,
                                )
                            },
                            label = { Text(item.label) },
                            colors = NavigationBarItemDefaults.colors(
                                selectedIconColor = MaterialTheme.colorScheme.primary,
                                selectedTextColor = MaterialTheme.colorScheme.primary,
                                indicatorColor = MaterialTheme.colorScheme.primaryContainer,
                            ),
                        )
                    }
                }
            }
        },
    ) { padding ->
        NavHost(
            navController = navController,
            startDestination = Screen.Recipes.route,
            modifier = Modifier.padding(padding),
            enterTransition = { enterTransition },
            exitTransition = { exitTransition },
            popEnterTransition = { popEnterTransition },
            popExitTransition = { popExitTransition },
        ) {
            composable(Screen.Recipes.route) {
                RecipeListScreen(
                    api = api,
                    serverUrl = serverUrl,
                    onRecipeClick = { id -> navController.navigate(Screen.RecipeDetail.create(id)) },
                    onImportClick = { navController.navigate(Screen.Import.route) },
                )
            }

            composable(Screen.Categories.route) {
                CategoriesScreen(
                    api = api,
                    onCategoryClick = { slug ->
                        navController.navigate(Screen.CategoryRecipes.create(slug))
                    },
                )
            }

            composable(
                Screen.CategoryRecipes.route,
                arguments = listOf(navArgument("slug") { type = NavType.StringType }),
            ) { backStackEntry ->
                val slug = backStackEntry.arguments?.getString("slug") ?: ""
                RecipeListScreen(
                    api = api,
                    serverUrl = serverUrl,
                    categorySlug = slug,
                    onRecipeClick = { id -> navController.navigate(Screen.RecipeDetail.create(id)) },
                    onImportClick = { navController.navigate(Screen.Import.route) },
                )
            }

            composable(
                Screen.TagRecipes.route,
                arguments = listOf(navArgument("slug") { type = NavType.StringType }),
            ) { backStackEntry ->
                val slug = backStackEntry.arguments?.getString("slug") ?: ""
                RecipeListScreen(
                    api = api,
                    serverUrl = serverUrl,
                    tagSlug = slug,
                    onRecipeClick = { id -> navController.navigate(Screen.RecipeDetail.create(id)) },
                    onImportClick = { navController.navigate(Screen.Import.route) },
                )
            }

            composable(Screen.MyRecipes.route) {
                MyRecipesScreen(
                    api = api,
                    serverUrl = serverUrl,
                    onRecipeClick = { id -> navController.navigate(Screen.RecipeDetail.create(id)) },
                )
            }

            composable(
                Screen.RecipeDetail.route,
                arguments = listOf(navArgument("id") { type = NavType.LongType }),
            ) { backStackEntry ->
                val id = backStackEntry.arguments?.getLong("id") ?: 0
                RecipeDetailScreen(
                    recipeId = id,
                    api = api,
                    serverUrl = serverUrl,
                    currentUserId = currentUserId,
                    onEdit = { recipeId -> navController.navigate(Screen.RecipeEdit.create(recipeId)) },
                    onDeleted = {
                        navController.popBackStack(Screen.Recipes.route, inclusive = false)
                    },
                    onCategoryClick = { slug -> navController.navigate(Screen.CategoryRecipes.create(slug)) },
                    onTagClick = { slug -> navController.navigate(Screen.TagRecipes.create(slug)) },
                )
            }

            composable(Screen.RecipeCreate.route) {
                RecipeFormScreen(
                    recipeId = null,
                    api = api,
                    onBack = { navController.popBackStack() },
                    onSaved = { id ->
                        navController.navigate(Screen.RecipeDetail.create(id)) {
                            popUpTo(Screen.Recipes.route)
                        }
                    },
                )
            }

            composable(
                Screen.RecipeEdit.route,
                arguments = listOf(navArgument("id") { type = NavType.LongType }),
            ) { backStackEntry ->
                val id = backStackEntry.arguments?.getLong("id") ?: 0
                RecipeFormScreen(
                    recipeId = id,
                    api = api,
                    onBack = { navController.popBackStack() },
                    onSaved = { recipeId ->
                        navController.navigate(Screen.RecipeDetail.create(recipeId)) {
                            popUpTo(Screen.Recipes.route)
                        }
                    },
                )
            }

            composable(Screen.Import.route) {
                ImportScreen(
                    api = api,
                    onBack = { navController.popBackStack() },
                    onImported = { id ->
                        navController.navigate(Screen.RecipeDetail.create(id)) {
                            popUpTo(Screen.Recipes.route)
                        }
                    },
                    onManualCreate = {
                        navController.navigate(Screen.RecipeCreate.route) {
                            popUpTo(Screen.Import.route) { inclusive = true }
                        }
                    },
                )
            }
        }
    }
}
