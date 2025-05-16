"use client";

import { useState, useEffect } from "react";
import { createClient } from "@/utils/supabase/client";
import { Card, CardContent } from "@/components/ui/card";
import { useCart } from "@/contexts/cart-context";

type Product = {
  id: number;
  name: string;
  price: number;
  category_id: number;
  available: boolean;
};

type Category = {
  id: number;
  name: string;
};

type GroupedProducts = {
  [category: string]: Product[];
};

// Color mapping for different categories
const categoryColors: Record<string, string> = {
  Drinks: "bg-blue-600/10",
  Foods: "bg-red-600/10",
  Sweets: "bg-yellow-600/10",
};

export default function ProductList() {
  const [groupedProducts, setGroupedProducts] =
    useState<GroupedProducts | null>(null);
  const [loading, setLoading] = useState(true);
  const supabase = createClient();
  const { addItem } = useCart();

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        // Fetch both products and categories
        const [productsResponse, categoriesResponse] = await Promise.all([
          supabase.from("products").select("*"),
          supabase.from("categories").select("*"),
        ]);

        if (productsResponse.error) {
          throw productsResponse.error;
        }

        if (categoriesResponse.error) {
          throw categoriesResponse.error;
        }

        const products = productsResponse.data as Product[];
        const categories = categoriesResponse.data as Category[];

        // Create a categories lookup map for easy access
        const categoryMap = categories.reduce(
          (map, category) => {
            map[category.id] = category.name;
            return map;
          },
          {} as Record<number, string>
        );

        // Group products by category name using the category_id
        const grouped = products.reduce(
          (acc: GroupedProducts, product: Product) => {
            const categoryName =
              categoryMap[product.category_id] || "Uncategorized";
            if (!acc[categoryName]) {
              acc[categoryName] = [];
            }
            acc[categoryName].push(product);
            return acc;
          },
          {}
        );

        // Sort products in each category by availability (available first) and then by name
        Object.keys(grouped).forEach(category => {
          grouped[category].sort((a, b) => {
            // First sort by availability
            if (a.available !== b.available) {
              return a.available ? -1 : 1; // Available products first
            }
            // Then sort by name
            return a.name.localeCompare(b.name);
          });
        });

        setGroupedProducts(grouped);
      } catch (error) {
        console.error("Error fetching products:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, []);

  const handleAddToCart = (product: Product) => {
    addItem({
      id: product.id,
      name: product.name,
      price: product.price,
      quantity: 1,
    });
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-pulse text-lg">Produkte werden geladen...</div>
      </div>
    );
  }

  if (!groupedProducts || Object.keys(groupedProducts).length === 0) {
    return (
      <div className="text-center py-12">
        <h3 className="text-xl font-medium">Keine Produkte verfügbar</h3>
        <p className="text-muted-foreground mt-2">
          Schauen Sie später wieder vorbei für unsere Produktliste.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-16 pb-8">
      {Object.keys(groupedProducts).sort().map((category) => (
        <div key={category} className="space-y-4">
          <h2 className="text-2xl font-medium">{category}</h2>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {groupedProducts[category].map((product) => (
              <Card
                key={product.id}
                className={`overflow-hidden h-full transition-shadow hover:shadow-md flex items-center ${
                  categoryColors[category] || categoryColors.default
                } ${product.available ? "" : "opacity-40 pointer-events-none"}`}
                onClick={() => handleAddToCart(product)}
                aria-label={`${product.name} zum Warenkorb hinzufügen`}
                role="button"
                tabIndex={0}
              >
                <CardContent className="p-4 flex flex-col w-full">
                  <div>
                    <h3 className="text-lg font-medium">{product.name}</h3>
                    <div className="flex flex-row justify-start items-center mt-1 gap-2">
                      <p className="text-sm text-muted-foreground">
                        CHF {product.price.toFixed(2)}
                      </p>

                      {!product.available && (
                        <>
                          <p className="text-sm text-muted-foreground">
                            &bull;
                          </p>
                          <p className="text-sm text-muted-foreground">
                            Nicht verfügbar
                          </p>
                        </>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
