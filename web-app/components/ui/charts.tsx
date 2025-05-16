"use client";

import * as React from "react";
import {
  AreaChart as RechartsAreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
  TooltipProps,
  BarChart as RechartsBarChart,
  Bar,
  PieChart as RechartsPieChart,
  Pie,
  Cell,
} from "recharts";
import { cn } from "@/lib/utils";

// Shared types for charts
interface ChartProps {
  data: any[];
  index: string;
  className?: string;
}

// Custom tooltip component with shadcn UI styling
const CustomTooltip = ({ active, payload, label }: TooltipProps<any, any>) => {
  if (!active || !payload || payload.length === 0) {
    return null;
  }

  return (
    <div className="rounded-lg border bg-background p-2 shadow-sm">
      <div className="grid grid-cols-2 gap-2">
        <div className="flex flex-col">
          <span className="text-[0.70rem] uppercase text-muted-foreground">
            {label}
          </span>
        </div>
        {payload.map((item, index) => (
          <div key={`item-${index}`} className="flex flex-col gap-1">
            <div className="flex items-center gap-1">
              <div
                className="h-1 w-1 rounded-full"
                style={{ background: item.color }}
              />
              <span className="text-[0.70rem] uppercase text-muted-foreground">
                {item.name}
              </span>
            </div>
            <span className="font-medium">
              {item.value}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};

// Area Chart
interface AreaChartProps extends ChartProps {
  categories: string[];
  colors?: string[];
  valueFormatter?: (value: number) => string;
  yAxisWidth?: number;
  showLegend?: boolean;
}

export function AreaChart({
  data,
  index,
  categories,
  colors = ["#3b82f6", "#10b981", "#f97316", "#8b5cf6", "#ec4899"],
  valueFormatter = (value) => value.toString(),
  yAxisWidth = 40,
  showLegend = false,
  className,
}: AreaChartProps) {
  return (
    <ResponsiveContainer width="100%" height="100%" className={cn(className)}>
      <RechartsAreaChart
        data={data}
        margin={{
          top: 10,
          right: 15,
          left: 5,
          bottom: 5,
        }}
      >
        <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
        <XAxis
          dataKey={index}
          tickLine={false}
          axisLine={false}
          tick={{ fontSize: 12 }}
          dy={10}
        />
        <YAxis
          width={yAxisWidth}
          tickLine={false}
          axisLine={false}
          tick={{ fontSize: 12 }}
          tickFormatter={valueFormatter}
        />
        <Tooltip content={<CustomTooltip />} />
        {showLegend && <Legend wrapperStyle={{ fontSize: "12px" }} />}
        {categories.map((category, i) => (
          <Area
            key={category}
            type="monotone"
            dataKey={category}
            fill={colors[i % colors.length]}
            stroke={colors[i % colors.length]}
            strokeWidth={2}
            fillOpacity={0.1}
            activeDot={{ r: 6 }}
          />
        ))}
      </RechartsAreaChart>
    </ResponsiveContainer>
  );
}

// Bar Chart
interface BarChartProps extends ChartProps {
  categories: string[];
  colors?: string[];
  valueFormatter?: (value: number) => string;
  yAxisWidth?: number;
  showLegend?: boolean;
}

export function BarChart({
  data,
  index,
  categories,
  colors = ["#3b82f6", "#10b981", "#f97316", "#8b5cf6", "#ec4899"],
  valueFormatter = (value) => value.toString(),
  yAxisWidth = 40,
  showLegend = false,
  className,
}: BarChartProps) {
  return (
    <ResponsiveContainer width="100%" height="100%" className={cn(className)}>
      <RechartsBarChart
        data={data}
        margin={{
          top: 10,
          right: 15,
          left: 5,
          bottom: 5,
        }}
      >
        <CartesianGrid strokeDasharray="3 3" opacity={0.2} vertical={false} />
        <XAxis
          dataKey={index}
          tickLine={false}
          axisLine={false}
          tick={{ fontSize: 12 }}
          dy={10}
        />
        <YAxis
          width={yAxisWidth}
          tickLine={false}
          axisLine={false}
          tick={{ fontSize: 12 }}
          tickFormatter={valueFormatter}
        />
        <Tooltip content={<CustomTooltip />} />
        {showLegend && <Legend wrapperStyle={{ fontSize: "12px" }} />}
        {categories.map((category, i) => (
          <Bar
            key={category}
            dataKey={category}
            fill={colors[i % colors.length]}
            radius={[4, 4, 0, 0]}
          />
        ))}
      </RechartsBarChart>
    </ResponsiveContainer>
  );
}

// Pie Chart
interface PieChartProps extends ChartProps {
  category: string;
  colors?: string[];
  valueFormatter?: (value: number) => string;
}

export function PieChart({
  data,
  index,
  category,
  colors = ["#3b82f6", "#10b981", "#f97316", "#8b5cf6", "#ec4899"],
  valueFormatter = (value) => value.toString(),
  className,
}: PieChartProps) {
  // Custom renderer for the pie chart labels
  const renderCustomizedLabel = ({
    cx,
    cy,
    midAngle,
    innerRadius,
    outerRadius,
    percent,
  }: any) => {
    const RADIAN = Math.PI / 180;
    const radius = innerRadius + (outerRadius - innerRadius) * 0.5;
    const x = cx + radius * Math.cos(-midAngle * RADIAN);
    const y = cy + radius * Math.sin(-midAngle * RADIAN);

    return (
      <text
        x={x}
        y={y}
        fill="white"
        textAnchor={x > cx ? "start" : "end"}
        dominantBaseline="central"
        fontSize={12}
      >
        {`${(percent * 100).toFixed(0)}%`}
      </text>
    );
  };

  return (
    <ResponsiveContainer width="100%" height="100%" className={cn(className)}>
      <RechartsPieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          labelLine={false}
          label={renderCustomizedLabel}
          outerRadius="80%"
          dataKey={category}
          nameKey={index}
        >
          {data.map((entry, index) => (
            <Cell
              key={`cell-${index}`}
              fill={colors[index % colors.length]}
              stroke="transparent"
            />
          ))}
        </Pie>
        <Tooltip
          formatter={(value) => [valueFormatter(value as number), ""]}
          contentStyle={{
            borderRadius: "8px",
            padding: "8px",
          }}
        />
        <Legend
          layout="horizontal"
          verticalAlign="bottom"
          align="center"
          iconSize={10}
          iconType="circle"
          wrapperStyle={{ fontSize: "12px", paddingTop: "10px" }}
        />
      </RechartsPieChart>
    </ResponsiveContainer>
  );
}