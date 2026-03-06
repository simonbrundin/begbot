// Shared helpers for Cucumber steps
import type { SupabaseClient } from "@supabase/supabase-js";

export async function realSignInFactory(
  supabaseUrl: string,
  supabaseKey: string,
) {
  // dynamic import to avoid pulling supabase in mock mode
  const { createClient } = await import("@supabase/supabase-js");
  const client: SupabaseClient = createClient(supabaseUrl, supabaseKey) as any;
  return async (email: string, password: string) => {
    return client.auth.signInWithPassword({ email, password });
  };
}

export async function requireRealSignIn(
  supabaseUrl: string | undefined,
  supabaseKey: string | undefined,
) {
  if (!supabaseUrl || !supabaseKey)
    throw new Error(
      "SUPABASE_URL and SUPABASE_KEY must be set to use real auth",
    );
  return realSignInFactory(supabaseUrl, supabaseKey);
}
