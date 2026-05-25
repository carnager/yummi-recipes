package com.yummi.app.data.api

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class LoginRequest(
    val username: String,
    val password: String,
)

@Serializable
data class RegisterRequest(
    val username: String,
    @SerialName("display_name") val displayName: String,
    val password: String,
)

@Serializable
data class AuthResponse(
    val token: String,
    val user: ApiUser,
)

@Serializable
data class ApiUser(
    val id: Long = 0,
    val username: String = "",
    @SerialName("display_name") val displayName: String = "",
)

@Serializable
data class ApiRecipe(
    @SerialName("ID") val id: Long = 0,
    @SerialName("Title") val title: String = "",
    @SerialName("Description") val description: String = "",
    @SerialName("SourceURL") val sourceUrl: String = "",
    @SerialName("ImagePath") val imagePath: String = "",
    @SerialName("PrepTime") val prepTime: String = "",
    @SerialName("CookTime") val cookTime: String = "",
    @SerialName("Servings") val servings: String = "",
    @SerialName("ContentMD") val contentMd: String = "",
    @SerialName("CategoryID") val categoryId: Long? = null,
    @SerialName("CreatedBy") val createdBy: Long = 0,
    @SerialName("CreatedAt") val createdAt: String = "",
    @SerialName("UpdatedAt") val updatedAt: String = "",
    @SerialName("Category") val category: ApiCategory? = null,
    @SerialName("Tags") val tags: List<ApiTag>? = null,
    @SerialName("AuthorName") val authorName: String = "",
    @SerialName("Tried") val tried: Boolean = false,
    @SerialName("SharedByName") val sharedByName: String = "",
)

@Serializable
data class ApiCategory(
    @SerialName("ID") val id: Long,
    @SerialName("Name") val name: String,
    @SerialName("Slug") val slug: String,
)

@Serializable
data class ApiTag(
    @SerialName("ID") val id: Long,
    @SerialName("Name") val name: String,
    @SerialName("Slug") val slug: String,
)

@Serializable
data class ImportRequest(val url: String)

@Serializable
data class CreateRecipeRequest(
    val title: String,
    val description: String = "",
    @SerialName("source_url") val sourceUrl: String = "",
    @SerialName("prep_time") val prepTime: String = "",
    @SerialName("cook_time") val cookTime: String = "",
    val servings: String = "",
    @SerialName("content_md") val contentMd: String = "",
    @SerialName("category_id") val categoryId: Long? = null,
    val tags: String = "",
)

@Serializable
data class TriedResponse(val tried: Boolean)

@Serializable
data class ShareRequest(
    @SerialName("user_id") val userId: Long,
    val action: String = "",
)

@Serializable
data class MyRecipesResponse(
    val own: List<ApiRecipe>? = null,
    val shared: List<ApiRecipe>? = null,
)

@Serializable
data class ApiError(val error: String)
