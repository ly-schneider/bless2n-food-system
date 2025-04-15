"use client";

import { useEffect, useState } from "react";
import { createClient } from "@/utils/supabase/client";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { formatDistanceToNow } from "date-fns";
import { de } from "date-fns/locale";
import { AreaChart, BarChart } from "../ui/charts";

type OrderSummary = {
  date: string;
  income: number;
  orders: number;
  products_sold: number;
};

type ProductTrend = {
  name: string;
  count: number;
};

// Matches Supabase's return; product is an array.
type SupabaseTrendingItem = {
  quantity: number;
  product: { name: string }[];
};

// Your desired type: product as a single object.
type TrendingItem = {
  quantity: number;
  product: { name: string };
};

export function StatisticsPanel() {
  const [dailyData, setDailyData] = useState<OrderSummary[]>([]);
  const [trendingProducts, setTrendingProducts] = useState<ProductTrend[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const supabase = createClient();

  const fetchData = async () => {
    setIsLoading(true);

    try {
      // Get the last 7 days data
      const sevenDaysAgo = new Date();
      sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);

      // Fetch daily revenue, orders, and products sold
      const { data: orderData, error: orderError } = await supabase
        .from("orders")
        .select(
          `
          id,
          order_date,
          total,
          order_items(quantity)
        `
        )
        .gte("order_date", sevenDaysAgo.toISOString());

      if (orderError) throw orderError;

      // Process data by day
      const dailySummary: Record<string, OrderSummary> = {};

      orderData.forEach((order) => {
        const date = new Date(order.order_date).toISOString().split("T")[0];

        if (!dailySummary[date]) {
          dailySummary[date] = {
            date,
            income: 0,
            orders: 0,
            products_sold: 0,
          };
        }

        dailySummary[date].income += Number(order.total);
        dailySummary[date].orders += 1;

        // Count products sold
        const itemsQuantity = order.order_items.reduce(
          (sum: number, item: { quantity: number }) => sum + item.quantity,
          0
        );
        dailySummary[date].products_sold += itemsQuantity;
      });

      const sortedDailyData = Object.values(dailySummary).sort(
        (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
      );

      setDailyData(sortedDailyData);

      const { data: rawTrendingData, error: trendingError } = (await supabase
        .from("order_items")
        .select("quantity, product:product_id(name)")
        .gte("created_at", sevenDaysAgo.toISOString())) as unknown as {
        data: SupabaseTrendingItem[];
        error: any;
      };

      if (trendingError) throw trendingError;
      if (!rawTrendingData) {
        throw new Error("No trending data found");
      }

      const trendingData: TrendingItem[] = rawTrendingData.map((item) => {
        let product;
        if (Array.isArray(item.product)) {
          // If it's an array, use the first element if available
          product =
            item.product.length > 0 ? item.product[0] : { name: "Unknown" };
        } else if (item.product) {
          // If it's already an object, use it directly
          product = item.product;
        } else {
          // Fallback case if product is null or undefined
          product = { name: "Unknown" };
        }
        return {
          quantity: item.quantity,
          product,
        };
      });

      const productCounts: Record<string, number> = {};

      trendingData.forEach((item) => {
        console.log(item);
        const productName = item.product.name;
        if (!productCounts[productName]) {
          productCounts[productName] = 0;
        }
        productCounts[productName] += item.quantity;
      });

      const topProducts = Object.entries(productCounts)
        .map(([name, count]) => ({ name, count }))
        .sort((a, b) => b.count - a.count)
        .slice(0, 3);

      setTrendingProducts(topProducts);
      setLastUpdated(new Date());
    } catch (error) {
      console.error("Error fetching statistics:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();

    // Setup refresh interval (every 5 minutes)
    const interval = setInterval(fetchData, 5 * 60 * 1000);

    return () => clearInterval(interval);
  }, []);

  // Format data for charts
  const formatChartData = () => {
    // Format daily data
    const formattedDailyData = dailyData.map((day) => ({
      name: new Date(day.date).toLocaleDateString("en-CH", {
        day: "numeric",
        month: "short",
      }),
      Revenue: Number(day.income.toFixed(2)),
      Products: day.products_sold,
      Orders: day.orders,
    }));

    // Format trending products data
    const formattedTrendingData = trendingProducts.map((product) => ({
      name: product.name,
      value: product.count,
    }));

    return {
      dailyData: formattedDailyData,
      trendingData: formattedTrendingData,
    };
  };

  if (isLoading) {
    return (
      <div className="w-full h-[400px] flex items-center justify-center">
        <p className="text-lg text-muted-foreground">Statistik wird geladen...</p>
      </div>
    );
  }

  const { dailyData: formattedDailyData, trendingData: formattedTrendingData } =
    formatChartData();

  return (
    <div className="space-y-8">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-medium">Dashboard-Übersicht</h2>
        {lastUpdated && (
          <p className="text-sm text-muted-foreground">
            Letztes Update:{" "}
            {formatDistanceToNow(lastUpdated, { addSuffix: true, locale: de })}
          </p>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Daily Revenue Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Tagesumsatz (CHF)</CardTitle>
            <CardDescription>
              Bruttoeinnahmen pro Tag der letzten 7 Tage
            </CardDescription>
          </CardHeader>
          <CardContent className="h-[350px]">
            <AreaChart
              data={formattedDailyData}
              categories={["Revenue"]}
              index="name"
              colors={["blue"]}
              valueFormatter={(value) => `CHF ${value}`}
              yAxisWidth={60}
              showLegend={true}
            />
          </CardContent>
        </Card>

        {/* Products Sold & Orders Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Produkte & Bestellungen</CardTitle>
            <CardDescription>
              Anzahl verkaufter Produkte und aufgegebener Bestellungen
            </CardDescription>
          </CardHeader>
          <CardContent className="h-[350px]">
            <BarChart
              data={formattedDailyData}
              categories={["Products", "Orders"]}
              index="name"
              colors={["green", "orange"]}
              yAxisWidth={40}
              showLegend={true}
            />
          </CardContent>
        </Card>

        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle>Top 3 Trendprodukte</CardTitle>
            <CardDescription>
              Beliebteste Produkte der letzten 7 Tage
            </CardDescription>
          </CardHeader>
          <CardContent className="h-[350px]">
            {trendingProducts.length > 0 ? (
              <div className="mx-auto w-full max-w-md">
                <BarChart
                  data={formattedTrendingData}
                  categories={["value"]}
                  index="name"
                  colors={["blue", "green", "orange"]}
                  valueFormatter={(value) => `${value} Stück`}
                  yAxisWidth={60}
                  showLegend={false}
                  className="h-[300px]"
                />
              </div>
            ) : (
              <div className="text-center text-muted-foreground h-full flex items-center justify-center">
                Keine Produktdaten verfügbar
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
