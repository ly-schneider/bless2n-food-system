// Follow this setup guide to integrate the Deno language server with your editor:
// https://deno.land/manual/getting_started/setup_your_environment
// This enables autocomplete, go to definition, etc.

import { createClient } from 'https://esm.sh/@supabase/supabase-js@2.39.1'

const corsHeaders = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
}

interface OrderItem {
  product_name: string;
  quantity: number;
  price: number;
  color_hex: string;
  emoji: string;
}

interface Order {
  id: string;
  items: OrderItem[];
  total: number;
  created_at: string;
}

Deno.serve(async (req) => {
  // Handle CORS preflight requests
  if (req.method === 'OPTIONS') {
    return new Response('ok', { headers: corsHeaders })
  }

  try {
    const supabaseClient = createClient(
      Deno.env.get('SUPABASE_URL') ?? '',
      Deno.env.get('SUPABASE_ANON_KEY') ?? '',
      {
        auth: {
          persistSession: false,
        },
      }
    )

    // Get order_id from request body if provided
    const { order_id } = await req.json().catch(() => ({}))

    if (!order_id) {
      throw new Error('order_id is required')
    }

    // Get all items in the order with their details
    const { data: orderItems, error: itemsError } = await supabaseClient
      .from('order_fulfillment')
      .select(`
        item_id,
        order_id,
        item:item (
          name,
          price,
          color_hex,
          emoji
        )
      `)
      .eq('order_id', order_id)

    if (itemsError) throw itemsError

    // Group items and count quantities
    const itemGroups = orderItems.reduce((acc: { [key: string]: OrderItem }, curr: any) => {
      const itemId = curr.item_id
      if (!acc[itemId]) {
        acc[itemId] = {
          product_name: curr.item.name,
          quantity: 1,
          price: curr.item.price,
          color_hex: curr.item.color_hex,
          emoji: curr.item.emoji
        }
      } else {
        acc[itemId].quantity++
      }
      return acc
    }, {})

    // Transform grouped items into array
    const items = Object.values(itemGroups)

    // Calculate total
    const total = items.reduce((sum, item) => sum + (item.quantity * item.price), 0)

    const order: Order = {
      id: order_id,
      items,
      total,
      created_at: new Date().toISOString() // You might want to get this from the order table if available
    }

    return new Response(
      JSON.stringify(order),
      {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        status: 200,
      },
    )
  } catch (error) {
    return new Response(
      JSON.stringify({ error: error.message }),
      {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        status: 400,
      },
    )
  }
})
