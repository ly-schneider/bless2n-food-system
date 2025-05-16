// Follow this setup guide to integrate the Deno language server with your editor:
// https://deno.land/manual/getting_started/setup_your_environment
// This enables autocomplete, go to definition, etc.

// Setup type definitions for built-in Supabase Runtime APIs
import "jsr:@supabase/functions-js/edge-runtime.d.ts"

// Import and initialize Supabase client using environment variables
import { createClient } from "@supabase/supabase-js";
import { corsHeaders } from '../_shared/cors.ts'

console.log(`Function "get_item" up and running!`)

Deno.serve(async (req: Request) => {
  // This is needed if you're planning to invoke your function from a browser.
  console.log(req.method);
  if (req.method === 'OPTIONS') {
    return new Response('ok', { headers: corsHeaders })
  }

  try {
    // Create a Supabase client with the Auth context of the logged in user.
    const supabaseClient = createClient(
      // Supabase API URL - env var exported by default.
      Deno.env.get('SUPABASE_URL') ?? '',
      // Supabase API ANON KEY - env var exported by default.
      Deno.env.get('SUPABASE_ANON_KEY') ?? '',
      // Create client with Auth context of the user that called the function.
      // This way your row-level-security (RLS) policies are applied.
      /*{
        global: {
          headers: { Authorization: req.headers.get('Authorization')! },
        },
      }*/
    )

    /*// First get the token from the Authorization header
    const token = req.headers.get('Authorization').replace('Bearer ', '')

    // Now we can get the session or user object
    const {
      data: { user },
    } = await supabaseClient.auth.getUser(token)*/

    const { id } = req.query; // Optional: Artikel-ID
    console.log("Now get the items");
    try {
      let data;
      if (id) {
        // Spezifischen Artikel abrufen
        const { data: item, error } = await supabase
          .from('item')
          .select('*')
          .eq('item_id', id)
          .single();
        
        if (error) throw error;
        data = item;
      } else {
        // Alle Artikel abrufen
        const { data: item, error } = await supabase
          .from('item')
          .select('*');
        
        if (error) throw error;
        data = item;
      }

      return new Response(JSON.stringify({ success: true, data }), { status: 200 });

    } catch (error) {
      console.error("Unexpected error:", error);
      return new Response(
        JSON.stringify({ error: "An unexpected error occurred." }),
        { headers: { ...corsHeaders, "Content-Type": "application/json" }, status: 500 },
      );
    }

  } catch (error) {
    return new Response(JSON.stringify({ error: error.message }), {
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
      status: 400,
    })
  }
})