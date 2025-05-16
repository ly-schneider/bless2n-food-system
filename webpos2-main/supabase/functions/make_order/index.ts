// Follow this setup guide to integrate the Deno language server with your editor:
// https://deno.land/manual/getting_started/setup_your_environment
// This enables autocomplete, go to definition, etc.

// Setup type definitions for built-in Supabase Runtime APIs
import "jsr:@supabase/functions-js/edge-runtime.d.ts"

// Import and initialize Supabase client using environment variables
import { createClient } from "@supabase/supabase-js";
import { corsHeaders } from '../_shared/cors.ts'

// Initialize the Supabase client

const supabase = createClient(
  Deno.env.get('SUPABASE_URL')!, // Supabase project URL
  Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')! // Supabase service role key
);

console.log("Supabase Function Initialized");

Deno.serve(async (req) => {
  // Handle CORS preflight requests (OPTIONS method)
  if (req.method === 'OPTIONS') {
    return new Response('ok', {
      headers: {
        ...corsHeaders,
        'Access-Control-Max-Age': '86400', // Cache preflight response for 1 day
      },
      status: 200,
    });
  }

  try {
    // Create a Supabase client with the Auth context of the logged in user.
    const supabaseClient = createClient(
      // Supabase API URL - env var exported by default.
      Deno.env.get('SUPABASE_URL') ?? '',
      // Supabase API ANON KEY - env var exported by default.
      Deno.env.get('SUPABASE_SERVICE_ROLE_KEY') ?? '',
      // Create client with Auth context of the user that called the function.
      // This way your row-level-security (RLS) policies are applied.
      /*{
        global: {
          headers: { Authorization: req.headers.get('Authorization')! },
        },
      }*/
    )

    // First get the token from the Authorization header
    /*const token = req.headers.get('Authorization').replace('Bearer ', '')

    // Now we can get the session or user object
    const {
      data: { user },
    } = await supabase.auth.getUser(token)*/
    

    // Step 1: Input validation and fetching item details
    try {
      // Parse the incoming JSON body
      const jsonData = await req.json();
      const { user_id, donation_amount, total_amount, discount_code, payment_method, items } = jsonData;

      console.log(jsonData);

      //If no user id is provided it takes the one from info@bless2n.ch
      const final_user_id = user_id || '464ae981-cb34-4d05-a7b8-9a298804e341';

      // Check if the required fields are present in the JSON request
      if (!final_user_id) {
        return new Response(JSON.stringify({ error: 'Missing user_id' }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 });
      }

      if (!total_amount || typeof total_amount !== 'number' || total_amount <= 0) {
        return new Response(JSON.stringify({ error: 'Missing or invalid total_amount' }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 });
      }

      // Validate donation_amount if provided
      if (donation_amount !== undefined && (typeof donation_amount !== 'number' || donation_amount < 0)) {
        return new Response(JSON.stringify({ error: 'Invalid donation_amount, it should be a positive number or zero' }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 });
      }

      // Validate the 'items' array exists and has at least one item
      if (!Array.isArray(items) || items.length === 0) {
        return new Response(JSON.stringify({ error: 'Missing items array or it is empty' }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 });
      }

      // Validate payment_type if provided
      if (!payment_method) {
        return new Response(JSON.stringify({ error: 'Missing payment method.' }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 });
      }

      // Loop through each item in the 'items' array to ensure valid data
      for (const [index, item] of items.entries()) {
        const item_id = item.id;
        const quantity = item.quantity;

        // Check if the 'item_id' is provided and valid
        if (!item_id) {
          return new Response(JSON.stringify({ error: `Missing item_id for item at index ${index}` }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 });
        }

        // Check if the 'quantity' is provided, is a number, and is greater than 0
        if (typeof quantity !== 'number' || quantity <= 0) {
          return new Response(JSON.stringify({ error: `Invalid quantity for item at index ${index}, must be a positive number` }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 });
        }
      }

      // Step 1.1: Fetch payment method ID from the 'payment_method' table
      const { data: paymentMethodData, error: paymentMethodError } = await supabase
        .from('payment_method') // Replace with your actual payment method table name
        .select('payment_method_id') // Ensure this is the correct column for ID
        .eq('payment_code', payment_method) // Match with the payment method code from the JSON
        .single();

      if (paymentMethodError || !paymentMethodData) {
        return new Response(
          JSON.stringify({ error: 'Invalid payment method or not found' }),
          { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 }
        );
      }

      // Retrieve the payment method ID for insertion
      const paymentMethodId = paymentMethodData.payment_method_id;
      console.log("Step 1");
      // Step 1.2: Fetch item details from the 'items' table and ensure enough stock is available for each item in the order
      const itemIds = Array.from(new Set(items.map((item: any) => item.id))); // Extract all item IDs from the order request
      
      const { data: itemData, error: itemError } = await supabase
        .from('item')
        .select('item_id, stock, price,name,type') // Select stock and price for each item
        .in('item_id', itemIds); // Get only the items in the current order
        
      // If there is an error or if not all items are found, return an error response
      if (itemError || itemData.length == 0) {
        return new Response(JSON.stringify({ error: 'Invalid items or stock error' }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 });
      }


      // Step 2: Calculate the order line amounts, check stock availability, and prepare data for the order lines
      let subtotal = 0; // Subtotal for all items before discounts and donations
      const orderLines: {
        item_id: number;
        quantity: number;
        price: number;
        line_amount: number;
        description: string;
      }[] = []; // Array to store individual order line data
      for (const orderItem of items) {
        // Find the item details in the fetched itemData using item_id
        const item = itemData.find((i: MenuItem) => i.item_id === orderItem.id);
        // If item not found, return an error
        if (!item) {
          return new Response(JSON.stringify({ error: `Item with ID ${orderItem.id} not found` }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 });
        }
        
        // Check stock after confirming item exists
        if (item.stock < orderItem.quantity) {
          return new Response(JSON.stringify({ error: `Insufficient stock for item ${item.name}` }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 });
        }
        
        // Calculate the total amount for this line (price * quantity)
        const lineAmount = item.price * orderItem.quantity;
        subtotal += lineAmount; // Add this line's total to the subtotal

        // Prepare the order line data to be inserted into the `order_line` table later
        orderLines.push({
          item_id: orderItem.item_id,
          quantity: orderItem.quantity,
          price: item.price,
          line_amount: lineAmount, // Total amount for this line
          description: item.name,
        });
      }

      // Step 3: Apply discount if a discount code is provided
      let discountAmount = 0;
      if (discount_code) {
        // Fetch discount details from the 'discount' table using the provided discount code
        const { data: discountData, error: discountError } = await supabase
          .from('discount')
          .select('discount_value, type, available_from, available_to, usage_limit') // Fetch necessary discount fields
          .eq('discount_code', discount_code) // Match the discount by code
          .single(); // Ensure we get only one discount record

        // If discount code is invalid or discount does not exist, return an error
        if (discountError || !discountData) {
          return new Response(JSON.stringify({ error: 'Invalid or expired discount code' }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 });
        }

        // Check if the discount is within its valid date range
        const now = new Date();
        if (now < new Date(discountData.available_from) || now > new Date(discountData.available_to)) {
          return new Response(JSON.stringify({ error: 'Discount code is not valid at this time' }), { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 });
        }

        // Apply the discount based on its type (percentage or fixed amount)
        if (discountData.type === 'Percent') {
          discountAmount = (subtotal * discountData.discount_value) / 100; // Apply percentage discount
        } else if (discountData.type === 'Amount') {
          discountAmount = discountData.discount_value; // Apply fixed amount discount
        }

        // (Optional) If you need to track usage limit, you can add logic here
        // Example: If the discount has a usage limit, you can check how many times it has been used and ensure it's not exceeded
        if (discountData.usage_limit) {
          // Logic to check usage limit can be implemented here
        }
      }
      

      // Final amount is the subtotal minus discount plus donation
      const finalAmount = subtotal - discountAmount + donation_amount;

      // Step 4: Validate the total amount by checking if the final amount matches the `total_amount` provided in the request
      if (finalAmount !== total_amount) {
        return new Response(
          JSON.stringify({ error: 'Total amount does not match calculated total' }),
          { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 400 }
        );
      }

      // Step 5: Create the order in the `order` table

      // Get the 'open' status_id from the 'order_status' table
      const { data: statusData, error: statusError } = await supabase
      .from('order_status')
      .select('status_id')
      .eq('status_name', 'open')
      .single();

      if (statusError || !statusData) {
        return new Response(
          JSON.stringify({ error: 'Failed to fetch status for order fulfillment.' }),
          { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 500 }
        );
      }
      const status_id = statusData.status_id;

      // Step 5.1: Get the next order number
      const { data: orderNumbers, error: orderNumberError } = await supabase
        .from('order')
        .select('order_number')
        .order('order_number', { ascending: false })
        .limit(1)
        .single(); // Fetch the highest order number

      // Determine the next order number
      let nextOrderNumber = 1000; // Start from 1000
      if (orderNumbers) {
        nextOrderNumber = orderNumbers.order_number + 1; // Increment the highest order number found
      } 

      const { data: newOrder, error: orderError } = await supabase
        .from('order')
        .insert({
          order_number: nextOrderNumber, // The next Ordernumber
          user_id: final_user_id, // The user who placed the order
          donation_amount: donation_amount, // The donation amount, if any
          total_amount: total_amount, // The total amount including donation and discount
          discount_code: discount_code, // Optional discount code
          discount_amount: discountAmount, // Discount amount applied to the order
          order_date: new Date(), // Set the order date to the current date and time
          status_id: status_id, // Replace with the default status UUID
          payment_method_id: paymentMethodId, // Insert the payment method ID
        })
        .select(); // Return the inserted order

      // If the order creation fails, return an error
      if (orderError || !newOrder) {
        return new Response(
          JSON.stringify({ 
            error: 'Failed to create order', 
            details: orderError // Füge die Details des Fehlers hinzu
          }), 
          { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 500 }
        );
      }

      const orderId = newOrder[0].order_id; // Get the order ID for the newly created order
      const orderNr = nextOrderNumber;

      // Step 6: Insert order lines into the order_line table and return their IDs
      const orderLineIds = [];
      for (const item of items) {
        const currentItemData = itemData.find((i: MenuItem) => i.item_id === item.id);
        // Calculate line amount
        if (!currentItemData) {
          throw new Error(`Item with ID ${item.id} not found in menu`);
        }
        const lineAmount = currentItemData.price * item.quantity; // Assuming you have price in item
        console.log("currentItemData: ",currentItemData);
        // Insert into order_line table
        const { data: orderLineData, error: orderLineError } = await supabase
          .from('order_line')
          .insert({
            order_id: orderId, // Reference to the order ID
            item_id: item.id, // Reference to the item ID
            quantity: item.quantity, // Quantity ordered
            price: currentItemData.price, // Price of the item
            line_amount: lineAmount, // Use the calculated line amount
            description: currentItemData.name
          })
          .select('order_line_id'); // Select the newly generated order_line_id

        // If there is an error inserting order lines, return an error response
        if (orderLineError) {
          return new Response(
            JSON.stringify({ 
              error: 'Failed to create order lines', 
              details: item // Füge die Details des Fehlers hinzu
            }), 
            { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 500 }
          );
        }
        const orderLineIds: string[] = [];
        // Store the generated order_line_id
        orderLineIds.push(orderLineData[0].order_line_id); // Assuming it returns an array with one object
        
        // Step 6.1: Insert fulfillment records into 'order_fulfillment' table
        const item_id = item.id;
        const quantity = item.quantity;
        const orderLineId = orderLineData[0].order_line_id;
        console.log("orderLineId",orderLineId)

        // If the item is of type 'product', create one fulfillment line for each quantity
        if (currentItemData.type === 'Product') {
          for (let i = 0; i < quantity; i++) {

            // Now insert into the order_fulfillment table with the order_line_id
            const { error: fulfillmentError } = await supabase
              .from('order_fulfillment')
              .insert<OrderFulfillment>({
                order_id: orderId,          // The order ID from the created order
                item_id: item_id,           // The item ID
                order_line_id: orderLineId, // Reference to the order line ID
                status_id: status_id,       // The 'open' status ID
              });

            if (fulfillmentError) {
              return new Response(
                JSON.stringify({ error: `Failed to insert fulfillment record for item ${item_id}` }),
                { headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 500 }
              );
            }
          }
        }

        // If the item is of type 'menu', insert fulfillment records for each product in the menu
        if (currentItemData.type === 'Menu') {
          // Loop through each product in the menu and insert fulfillment records
          if (item.products && Array.isArray(item.products)) {
            for (const menuItem of item.products) {
              const product_id = menuItem.id;
              const product_quantity = menuItem.quantity * quantity; // Adjust for ordered quantity of menu

              for (let i = 0; i < product_quantity; i++) {
                
                // Now insert into the order_fulfillment table with the order_line_id
                const { error: fulfillmentError } = await supabase
                  .from('order_fulfillment')
                  .insert<OrderFulfillment>({
                    order_id: orderId,          // The order ID from the created order
                    item_id: product_id,        // The product ID from the menu
                    order_line_id: orderLineId, // Reference to the order line ID
                    status_id: status_id,       // The 'open' status ID
                  });

                if (fulfillmentError) {
                  return new Response(
                    JSON.stringify({ error: `Failed to insert fulfillment record for product ${product_id}` }),
                    {headers: { ...corsHeaders, 'Content-Type': 'application/json' }, status: 500 }
                  );
                }
              }
            }
          }
        }


      }

      // Step 7: Update stock for each item in the `item` table
      for (const item of items) {
        // Fetch the current stock and type for the item
        const { data: currentItemData, error: currentItemError } = await supabase
          .from('item')
          .select('stock, type')
          .eq('item_id', item.id)
          .single();
        

        // If there's an error fetching the current stock, return an error response
        if (currentItemError || !currentItemData) {
          return new Response(JSON.stringify({ error: 'Failed to fetch item stock' }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' },
            status: 500
          });
        }

        const currentStock = currentItemData.stock; // Get current stock
        const itemType = currentItemData.type; // Get item type

        if (itemType === "Product") {
          // Process as a regular product
          if (currentStock < item.quantity) {
            return new Response(JSON.stringify({ error: 'Insufficient stock for item ' + item.name }), {
              headers: { ...corsHeaders, 'Content-Type': 'application/json' },
              status: 400
            });
          }

          // Calculate new stock
          const newStock = currentStock - item.quantity;

          // Update the stock for the product item
          const { error: stockUpdateError } = await supabase
            .from('item')
            .update({ stock: newStock })
            .eq('item_id', item.id);

          // If stock update fails, return an error
          if (stockUpdateError) {
            return new Response(JSON.stringify({ error: 'Failed to update stock' }), {
              headers: { ...corsHeaders, 'Content-Type': 'application/json' },
              status: 500
            });
          }

          // Verify the update
          const { data: verifyUpdate, error: verifyError } = await supabase
            .from('item')
            .select('stock')
            .eq('item_id', item.id)
            .single();

          if (verifyError || verifyUpdate.stock !== newStock) {
            console.error('Stock verification failed:', verifyError || 'Stock not updated correctly');
            return new Response(JSON.stringify({ 
              error: 'Stock update verification failed',
              details: verifyError || 'Stock not updated to expected value'
            }), {
              headers: { ...corsHeaders, 'Content-Type': 'application/json' },
              status: 500
            });
          }
        } else if (itemType === "Menu") {
          // Iterate through each product in the menu and update stock accordingly
          if (item.products && Array.isArray(item.products)) {
            for (const menuItem of item.products) {
              const product_id = menuItem.id;
              const quantity = menuItem.quantity;
              // Fetch the stock for each product in the menu
              const { data: productData, error: productError } = await supabase
                .from('item')
                .select('stock,name')
                .eq('item_id', product_id)
                .single();

              if (productError || !productData) {
                return new Response(JSON.stringify({ error: 'Failed to fetch product stock for menu item: '+menuItem.product_id }), {
                  headers: { ...corsHeaders, 'Content-Type': 'application/json' },
                  status: 500
                });
              }

              const productStock = productData.stock;

              // Check if there's enough stock for each product in the menu
              const requiredQuantity = quantity;
              if (productStock < requiredQuantity) {
                return new Response(JSON.stringify({ error: 'Insufficient stock for product in menu: ' + product_id }), {
                  headers: { ...corsHeaders, 'Content-Type': 'application/json' },
                  status: 400
                });
              }

              // Calculate new stock for the product
              const newProductStock = productStock - requiredQuantity;
              // Update the stock for each product in the menu
              const { error: productStockUpdateError } = await supabase
                .from('item')
                .update({ stock: newProductStock })
                .eq('item_id', product_id);

              // If product stock update fails, return an error
              if (productStockUpdateError) {
                return new Response(JSON.stringify({ error: 'Failed to update stock for product in menu' }), {
                  headers: { ...corsHeaders, 'Content-Type': 'application/json' },
                  status: 500
                });
              }
            }
          }
        }
      }
      // Step 9: Return a success message with the order ID
      return new Response(JSON.stringify({ message: 'Order created successfully', order_id: orderId, order_nr: orderNr }), {
        headers: { 
          ...corsHeaders, 
          'Content-Type': 'application/json',
        },
        status: 200, // Successful response
      });

    } catch (error) {
      // Catch any unexpected errors and return a 500 response
      return new Response(JSON.stringify({ error: error.message }), {
        headers: { 
        ...corsHeaders, 
        'Content-Type': 'application/json',
      }, 
      status: 500 });
    }
  } catch (error) {
    return new Response(JSON.stringify({ error: error.message }), {
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
      status: 400,
  })
}
});

// Interface definitions
interface MenuItem {
  item_id: string | number;
  type: string;
  stock?: number;
  price: number;
  name: string;
}

interface OrderItem {
  id: string | number;
  quantity: number;
}

interface OrderFulfillment {
  order_fulfillment_id?: string;  // UUID, auto-generated
  created_at?: string;            // Timestamp, auto-generated
  order_id: string;               // UUID
  item_id: string;                // UUID
  order_line_id: string;          // UUID
  status_id: string;              // UUID
}

/* To invoke locally:

  1. Run `supabase start` (see: https://supabase.com/docs/reference/cli/supabase-start)
  2. Make an HTTP request:

  curl -i --location --request POST 'http://127.0.0.1:54321/functions/v1/make_§' \
    --header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0' \
    --header 'Content-Type: application/json' \
    --data '{"name":"Functions"}'

*/
