export default defineNuxtRouteMiddleware((to) => {
  const user = useSupabaseUser()
  
  // List of routes that don't require authentication
  const publicRoutes = ['/login', '/confirm']
  
  // Check if the current route is public
  const isPublicRoute = publicRoutes.some(route => to.path === route)
  
  // If user is not authenticated and trying to access a protected route
  if (!user.value && !isPublicRoute) {
    return navigateTo('/login')
  }
  
  // If user is authenticated and trying to access login page, redirect to home
  if (user.value && to.path === '/login') {
    return navigateTo('/')
  }
})
