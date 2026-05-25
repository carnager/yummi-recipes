package com.yummi.app.ui.components

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.PersonAdd
import androidx.compose.material.icons.filled.PersonRemove
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.yummi.app.data.api.ApiUser
import com.yummi.app.data.api.ShareRequest
import com.yummi.app.data.api.YummiApi
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ShareBottomSheet(
    recipeId: Long,
    currentUserId: Long,
    api: YummiApi,
    onDismiss: () -> Unit,
) {
    var allUsers by remember { mutableStateOf<List<ApiUser>>(emptyList()) }
    var sharedWith by remember { mutableStateOf<Set<Long>>(emptySet()) }
    var isLoading by remember { mutableStateOf(true) }
    val scope = rememberCoroutineScope()

    LaunchedEffect(recipeId) {
        try {
            val usersResp = api.listUsers()
            if (usersResp.isSuccessful) {
                allUsers = (usersResp.body() ?: emptyList()).filter { it.id != currentUserId }
            }
        } catch (_: Exception) {}
        try {
            val sharesResp = api.listSharesForRecipe(recipeId)
            if (sharesResp.isSuccessful) {
                sharedWith = (sharesResp.body() ?: emptyList()).map { it.id }.toSet()
            }
        } catch (_: Exception) {}
        isLoading = false
    }

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        shape = RoundedCornerShape(topStart = 16.dp, topEnd = 16.dp),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .padding(bottom = 32.dp),
        ) {
            Text(
                text = "Rezept teilen",
                style = MaterialTheme.typography.headlineSmall,
                modifier = Modifier.padding(bottom = 16.dp),
            )

            if (isLoading) {
                Box(
                    modifier = Modifier.fillMaxWidth().height(100.dp),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
                }
            } else if (allUsers.isEmpty()) {
                Text(
                    text = "Keine anderen Benutzer vorhanden",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            } else {
                LazyColumn(
                    verticalArrangement = Arrangement.spacedBy(4.dp),
                    modifier = Modifier.heightIn(max = 400.dp),
                ) {
                    items(allUsers, key = { it.id }) { user ->
                        val isShared = user.id in sharedWith
                        UserShareRow(
                            user = user,
                            isShared = isShared,
                            onToggle = {
                                scope.launch {
                                    try {
                                        val action = if (isShared) "remove" else ""
                                        api.shareRecipe(recipeId, ShareRequest(userId = user.id, action = action))
                                        sharedWith = if (isShared) {
                                            sharedWith - user.id
                                        } else {
                                            sharedWith + user.id
                                        }
                                    } catch (_: Exception) {}
                                }
                            },
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun UserShareRow(
    user: ApiUser,
    isShared: Boolean,
    onToggle: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = user.displayName.ifBlank { user.username },
                style = MaterialTheme.typography.bodyLarge,
            )
            if (user.displayName.isNotBlank() && user.displayName != user.username) {
                Text(
                    text = "@${user.username}",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
        }

        IconButton(onClick = onToggle) {
            if (isShared) {
                Icon(
                    Icons.Default.PersonRemove,
                    contentDescription = "Teilen aufheben",
                    tint = MaterialTheme.colorScheme.error,
                )
            } else {
                Icon(
                    Icons.Default.PersonAdd,
                    contentDescription = "Teilen",
                    tint = MaterialTheme.colorScheme.primary,
                )
            }
        }
    }
    HorizontalDivider(color = MaterialTheme.colorScheme.outlineVariant)
}
