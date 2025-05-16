// Follow this setup guide to integrate the Deno language server with your editor:
// https://deno.land/manual/getting_started/setup_your_environment
// This enables autocomplete, go to definition, etc.

// Setup type definitions for built-in Supabase Runtime APIs
import "jsr:@supabase/functions-js/edge-runtime.d.ts"

// Import and initialize Supabase client using environment variables
import { createClient } from "@supabase/supabase-js";
import { corsHeaders } from '../_shared/cors.ts'

console.log(`Function "menucard_GET" up and running!`)

Deno.serve(async (req: Request) => {
  // This is needed if you're planning to invoke your function from a browser.
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

    try {
      // Step 3: Fetch all available products
      console.log("Fetching available products...");
      const { data: products, error: productError } = await supabaseClient
        .from('item')
        .select('item_id, name, stock, emoji, price, color_hex, type')
        .eq('status', 'Available')
        .eq('type', 'Product')
        .order('sequence', { ascending: true });
  
      if (productError) {
        console.error("Error fetching products:", productError);
        return new Response(
          JSON.stringify({ error: productError.message }),
          { headers: { ...corsHeaders, "Content-Type": "application/json" }, status: 500 },
        );
      }

      // Step 3.1: Fetch all available product Groups
      console.log("Fetching available product groups...");
      const { data: productGroups, error: productGroupError } = await supabaseClient
        .from('item')
        .select('item_id, name, stock, emoji, price, type,color_hex')
        .eq('status', 'Available')
        .eq('type', 'Product Group')
        .order('sequence', { ascending: true });
  
      if (productGroupError) {
        console.error("Error fetching products:", productGroupError);
        return new Response(
          JSON.stringify({ error: productGroupError.message }),
          { headers: { ...corsHeaders, "Content-Type": "application/json" }, status: 500 },
        );
      }

      console.log("Fetched products:", products);
  
      // Step 4: Fetch available menus
      console.log("Fetching available menus...");
      const { data: menus, error: menuError } = await supabaseClient
        .from('item')
        .select('item_id, name, stock, emoji, price')
        .eq('status', 'Available')
        .eq('type', 'Menu')
        .order('sequence', { ascending: true });
  
      if (menuError) {
        console.error("Error fetching menus:", menuError);
        return new Response(
          JSON.stringify({ error: menuError.message }),
          { headers: { ...corsHeaders, "Content-Type": "application/json" }, status: 500 },
        );
      }
      console.log("Fetched menus:", menus);
  
      // Step 5: Map through menus and fetch products for each menu
      console.log("Mapping through menus to fetch products...");
      const menusWithProducts = await Promise.all(menus.map(async (menu) => {
        console.log("Fetching menu items for menu ID:", menu.item_id);
        const { data: menuItems, error: menuItemError } = await supabaseClient
          .from('menuitems')
          .select('product_id, quantity')
          .eq('menu_id', menu.item_id)
          .eq('status', 'Available')
          .order('sequence', { ascending: true });
  
        if (menuItemError) {
          console.error("Error fetching menu items for menu ID:", menu.item_id, menuItemError);
          return null; // Or handle error as appropriate
        }
  
        console.log(`Fetched menu items for menu ID ${menu.item_id}:`, menuItems);
        
        // Step 6.1: Fetch and calculate stock based on products for each menu
        const { data:menuStock, error:errorMenuStock } = await supabaseClient.rpc('get_menu_stock', { par_menu_id: menu.item_id });
        if (errorMenuStock) {
            console.error('Error calling get_menu_stock:', errorMenuStock);
            return { error: errorMenuStock.message };
        }

        // Step 6: Construct the products array for each menu
        const productsForMenu = await Promise.all(menuItems.map(async (menuItem) => {
          const product = products.find(p => p.item_id === menuItem.product_id);
          const productGroup = productGroups.find(pg => pg.item_id === menuItem.product_id);
          
          if (product) {
            // If it's a product, return the product as usual
            return {
              id: product.item_id,
              name: product.name,
              type: product.type,
              emoji: product.emoji,
              quantity: menuItem.quantity,
              color: product.color_hex,
            };
          }
          
          if (productGroup) {
            // If it's a product group, fetch the products inside the group
            console.log(`Fetching products for product group ${productGroup.item_id}`);
            
            const { data: groupProducts, error: groupProductError } = await supabaseClient
              .from('menuitems')
              .select('product_id,quantity')
              .eq('menu_id', productGroup.item_id) 
              .eq('status', 'Available')
              .order('sequence', { ascending: true });

            if (groupProductError) {
              console.error("Error fetching products for group:", groupProductError);
              return null; // Or handle error as appropriate
            }

            // Fetch the products for the group
            const detailedProducts = await Promise.all(groupProducts.map(async (groupProduct) => {
              const detailedProduct = products.find(p => p.item_id === groupProduct.product_id);
              
              // Add a null check and handle the case where no product is found
              if (!detailedProduct) {
                console.warn(`No product found for product_id: ${groupProduct.product_id}`);
                return null; // or throw an error, depending on your requirements
              }

              return {
                id: detailedProduct.item_id,
                name: detailedProduct.name,
                emoji: detailedProduct.emoji,
                quantity: groupProduct.quantity,
                color: detailedProduct.color_hex,
              };
            })).then(products => products.filter(p => p !== null)); // Remove any null entries

            // Return the product group with the products nested inside
            return {
              id: productGroup.item_id,
              name: productGroup.name,
              type: productGroup.type,
              emoji: productGroup.emoji,
              quantity: menuItem.quantity,
              color: productGroup.color_hex,
              products: detailedProducts, // Nest the products inside the group
            };
          }

          return null;
        }));

  
        console.log(`Constructed products for menu ID ${menu.item_id}:`, productsForMenu);
  
        return {
          id: menu.item_id,
          name: menu.name,
          type: 'Menu',
          price: menu.price,
          stock: menuStock,
          emoji: menu.emoji,
          products: productsForMenu.filter(product => product !== null), // Filter out null products
        };
      }));
  
      // Step 7: Combine products and menus for the final response
      console.log("Combining products and menus for the final response...");
      const combinedResponse = [
        ...products.map(product => ({
          id: product.item_id,
          name: product.name,
          type: 'Product',
          price: product.price,
          stock: product.stock,
          emoji: product.emoji,
          color: product.color_hex
        })),
        ...menusWithProducts.filter(menu => menu !== null) // Filter out null menus
      ];
  
      console.log("Final combined response:", combinedResponse);
  
      return new Response(JSON.stringify(combinedResponse), {
        headers: {
          ...corsHeaders,
          "Content-Type": "application/json",
        },
      });
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

// To invoke:
// curl -i --location --request POST 'http://localhost:54321/functions/v1/select-from-table-with-auth-rls' \
//   --header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24ifQ.625_WdcF3KHqz5amU0x2X5WWHP-OEs_4qj0ssLNHzTs' \
//   --header 'Content-Type: application/json' \
//   --data '{"name":"Functions"}'