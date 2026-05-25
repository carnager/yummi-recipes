package com.yummi.app.data.api

import retrofit2.Response
import retrofit2.http.*

interface YummiApi {

    // Auth (no token needed)
    @POST("api/v1/auth/login")
    suspend fun login(@Body request: LoginRequest): Response<AuthResponse>

    @POST("api/v1/auth/register")
    suspend fun register(@Body request: RegisterRequest): Response<AuthResponse>

    // Recipes
    @GET("api/v1/recipes")
    suspend fun listRecipes(
        @Query("q") query: String? = null,
        @Query("category") category: String? = null,
        @Query("tag") tag: String? = null,
    ): Response<List<ApiRecipe>>

    @GET("api/v1/recipes/{id}")
    suspend fun getRecipe(@Path("id") id: Long): Response<ApiRecipe>

    @POST("api/v1/recipes")
    suspend fun createRecipe(@Body request: CreateRecipeRequest): Response<ApiRecipe>

    @PUT("api/v1/recipes/{id}")
    suspend fun updateRecipe(@Path("id") id: Long, @Body request: CreateRecipeRequest): Response<ApiRecipe>

    @DELETE("api/v1/recipes/{id}")
    suspend fun deleteRecipe(@Path("id") id: Long): Response<Unit>

    @POST("api/v1/recipes/import")
    suspend fun importRecipe(@Body request: ImportRequest): Response<ApiRecipe>

    @POST("api/v1/recipes/{id}/tried")
    suspend fun toggleTried(@Path("id") id: Long): Response<TriedResponse>

    @POST("api/v1/recipes/{id}/share")
    suspend fun shareRecipe(@Path("id") id: Long, @Body request: ShareRequest): Response<Unit>

    // My Recipes
    @GET("api/v1/my-recipes")
    suspend fun myRecipes(): Response<MyRecipesResponse>

    // Shares for a recipe
    @GET("api/v1/recipes/{id}/shares")
    suspend fun listSharesForRecipe(@Path("id") id: Long): Response<List<ApiUser>>

    // Categories & Tags
    @GET("api/v1/categories")
    suspend fun listCategories(): Response<List<ApiCategory>>

    @GET("api/v1/tags")
    suspend fun listTags(): Response<List<ApiTag>>

    // Users
    @GET("api/v1/users")
    suspend fun listUsers(): Response<List<ApiUser>>
}
