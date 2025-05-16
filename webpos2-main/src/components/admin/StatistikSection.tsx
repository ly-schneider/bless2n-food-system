'use client'

import { useState, useEffect } from 'react'
import { useSupabase } from '@/hooks/useSupabase'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { 
  AreaChart, 
  Area, 
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  Legend
} from 'recharts'
import { addDays, startOfDay, endOfDay, format, parseISO } from 'date-fns'
import { de } from 'date-fns/locale'
import { MultiDatePicker } from '@/components/ui/multi-date-picker'

interface OrderData {
  created_at: string
  total_amount: number
  payment_method: Array<{
    payment_method_id: string
    name: string
  }>
}

interface DailyRevenue {
  date: Date
  revenue: number
}

interface PaymentMethodStats {
  method: string
  amount: number
}

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8'];

const PAYMENT_METHOD_NAMES: { [key: string]: string } = {
  'CSH': 'Bargeld',
  'CRE': 'Kreditkarte',
  'TWI': 'Twint',
  'EMP': 'Mitarbeiter',
  'VIP': 'VIP',
  'KUL': 'Kultur',
  'GUT': 'Gutschein',
  'DIV': 'Diverses'
};

export function StatistikSection() {
  const { supabase } = useSupabase()
  const [selectedDates, setSelectedDates] = useState<Date[]>([])
  const [availableDates, setAvailableDates] = useState<Date[]>([])
  const [revenue, setRevenue] = useState<number>(0)
  const [dailyRevenue, setDailyRevenue] = useState<DailyRevenue[]>([])
  const [paymentMethodStats, setPaymentMethodStats] = useState<PaymentMethodStats[]>([])
  const [orderCount, setOrderCount] = useState<number>(0)
  const [averageOrderValue, setAverageOrderValue] = useState<number>(0)

  const fetchAvailableDates = async () => {
    try {
      const { data, error } = await supabase
        .from('order')
        .select('created_at')
        .order('created_at', { ascending: false })

      if (error) {
        console.error('Error fetching available dates:', error)
        return
      }

      if (!data || data.length === 0) {
        console.log('No orders found')
        return
      }

      // Convert to unique dates and sort in descending order
      const uniqueDates = Array.from(new Set(
        data.map(order => format(new Date(order.created_at), 'yyyy-MM-dd'))
      ))
        .map(dateStr => {
          const [year, month, day] = dateStr.split('-').map(Number)
          return new Date(year, month - 1, day)  // month is 0-based in JS Date
        })
        .sort((a, b) => b.getTime() - a.getTime())

      setAvailableDates(uniqueDates)
      
      // Select last 30 days by default
      const last30Days = uniqueDates.slice(0, 30)
      setSelectedDates(last30Days)
    } catch (error) {
      console.error('Error in fetchAvailableDates:', error)
    }
  }

  const fetchOrderData = async (dates: Date[]) => {
    if (!dates.length) return

    try {
      // Format dates for the query
      const dateConditions = dates.map(date => 
        `order_date.gte.${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}T00:00:00,order_date.lte.${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}T23:59:59.999`
      ).join(',')

      console.log('Fetching orders with conditions:', dateConditions)

      const { data: orders, error } = await supabase
        .from('order')
        .select(`
          created_at,
          order_date,
          total_amount,
          payment_method:payment_method_id(
            payment_method_id,
            description
          )
        `)
        .or(dateConditions)

      console.log('Supabase response:', { orders, error })

      if (error) {
        console.error('Error fetching orders:', error)
        return
      }

      if (!orders || orders.length === 0) {
        console.log('No orders found in the selected dates')
        setRevenue(0)
        setOrderCount(0)
        setAverageOrderValue(0)
        setDailyRevenue([])
        setPaymentMethodStats([])
        return
      }

      // Calculate total revenue and order count
      const totalRevenue = orders.reduce((sum, order) => sum + (order.total_amount || 0), 0)
      setRevenue(totalRevenue)
      setOrderCount(orders.length)
      setAverageOrderValue(totalRevenue / orders.length)

      // Process daily revenue
      const dailyData: { [key: string]: number } = {}
      orders.forEach((order: any) => {
        const orderDate = new Date(order.order_date)
        const dateKey = format(orderDate, 'yyyy-MM-dd')
        dailyData[dateKey] = (dailyData[dateKey] || 0) + (order.total_amount || 0)
      })

      const dailyRevenueData = Object.entries(dailyData).map(([dateStr, revenue]) => ({
        date: parseISO(dateStr),
        revenue
      })).sort((a, b) => a.date.getTime() - b.date.getTime())

      setDailyRevenue(dailyRevenueData)

      // Process payment methods
      const paymentData: { [key: string]: number } = {}
      orders.forEach((order: any) => {
        if (order.payment_method && order.payment_method.length > 0) {
          const methodName = order.payment_method[0].description || 'Unbekannt'
          paymentData[methodName] = (paymentData[methodName] || 0) + (order.total_amount || 0)
        }
      })

      const paymentStats = Object.entries(paymentData).map(([method, amount]) => ({
        method,
        amount
      })).sort((a, b) => b.amount - a.amount)

      setPaymentMethodStats(paymentStats)
    } catch (error) {
      console.error('Error in fetchOrderData:', error)
    }
  }

  useEffect(() => {
    fetchAvailableDates()
  }, [])

  useEffect(() => {
    if (selectedDates.length > 0) {
      fetchOrderData(selectedDates)
    } else {
      // Reset stats when no dates are selected
      setRevenue(0)
      setOrderCount(0)
      setAverageOrderValue(0)
      setDailyRevenue([])
      setPaymentMethodStats([])
    }
  }, [selectedDates, supabase])

  const handleDateChange = (newDates: Date[]) => {
    console.log('Dates changed to:', newDates)
    setSelectedDates(newDates)
  }

  return (
    <div className="space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Statistik</h2>
        <MultiDatePicker
          dates={selectedDates}
          setDates={handleDateChange}
          availableDates={availableDates}
          label="Zeitraum wählen"
        />
      </div>
      
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Gesamtumsatz</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'CHF' }).format(revenue)}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Anzahl Bestellungen</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {orderCount}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Durchschnittlicher Bestellwert</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'CHF' }).format(averageOrderValue)}
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Täglicher Umsatz</CardTitle>
          </CardHeader>
          <CardContent className="h-[400px]">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={dailyRevenue}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis 
                  dataKey="date"
                  tick={{ fontSize: 12 }}
                  tickFormatter={(date: Date) => format(date, 'dd.MM.yyyy')}
                  interval={Math.floor(dailyRevenue.length / 5)}
                />
                <YAxis 
                  tick={{ fontSize: 12 }}
                  tickFormatter={(value: number) => 
                    new Intl.NumberFormat('de-DE', { 
                      style: 'currency', 
                      currency: 'CHF',
                      minimumFractionDigits: 0,
                      maximumFractionDigits: 0
                    }).format(value)
                  }
                />
                <Tooltip 
                  formatter={(value: number) => 
                    new Intl.NumberFormat('de-DE', { 
                      style: 'currency', 
                      currency: 'CHF' 
                    }).format(value)
                  }
                />
                <Area 
                  type="monotone" 
                  dataKey="revenue" 
                  stroke="#8884d8" 
                  fill="#8884d8" 
                  name="Umsatz"
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Zahlungsmethoden</CardTitle>
          </CardHeader>
          <CardContent className="h-[400px]">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={paymentMethodStats}
                  dataKey="amount"
                  nameKey="method"
                  cx="50%"
                  cy="50%"
                  outerRadius={150}
                  label={({ method, amount }) => 
                    `${method}: ${new Intl.NumberFormat('de-DE', { 
                      style: 'currency', 
                      currency: 'CHF',
                      minimumFractionDigits: 0,
                      maximumFractionDigits: 0 
                    }).format(amount)}`
                  }
                >
                  {paymentMethodStats.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip 
                  formatter={(value: number) => 
                    new Intl.NumberFormat('de-DE', { 
                      style: 'currency', 
                      currency: 'CHF' 
                    }).format(value)
                  }
                />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
