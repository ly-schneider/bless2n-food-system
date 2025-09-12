"use client"

import { AlertCircle, Camera, CameraOff, CheckCircle, Hash, QrCode, Search } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Separator } from "@/components/ui/separator"
import RedemptionAPI from "@/lib/api/redemption"

type OrderItem = { name: string; quantity: number; price: number }
type OrderResult = {
  id: string
  orderNumber: string
  customer: { name: string; phone: string }
  status: string
  type: string
  items: OrderItem[]
  total: number
  qrCode: string
}
type TableResult = {
  id: string
  tableNumber: string
  capacity: number
  status: string
  currentOrder?: string
  qrCode: string
}

export default function ScanPage() {
  const [scanMode, setScanMode] = useState<"qr" | "manual">("qr")
  const [isScanning, setIsScanning] = useState(false)
  const [scanResult, setScanResult] = useState<string | null>(null)
  const [manualCode, setManualCode] = useState("")
  const [orderData, setOrderData] = useState<OrderResult | TableResult | null>(null)
  const [error, setError] = useState<string | null>(null)
  const videoRef = useRef<HTMLVideoElement>(null)

  // Real order lookup function
  const lookupOrder = async (code: string) => {
    try {
      const res = await RedemptionAPI.getOrderForRedemption(code)
      // Shape minimal fields for UI; backend may not include items/details
      const result: OrderResult = {
        id: res.id,
        orderNumber: code,
        customer: { name: res.contactEmail || "Customer", phone: "" },
        status: res.status,
        type: "takeout",
        items: [],
        total: res.total,
        qrCode: "",
      }
      return result
    } catch {
      return null
    }
  }

  const startCamera = async () => {
    try {
      setError(null)
      const stream = await navigator.mediaDevices.getUserMedia({
        video: {
          facingMode: "environment", // Use back camera if available
          width: { ideal: 1920 },
          height: { ideal: 1080 },
        },
      })

      if (videoRef.current) {
        videoRef.current.srcObject = stream
        setIsScanning(true)
      }
    } catch (err) {
      console.error("Camera access error:", err)
      setError("Camera access denied. Please enable camera permissions or use manual entry.")
      setScanMode("manual")
    }
  }

  const stopCamera = () => {
    if (videoRef.current?.srcObject) {
      const tracks = (videoRef.current.srcObject as MediaStream).getTracks()
      tracks.forEach((track) => track.stop())
      videoRef.current.srcObject = null
    }
    setIsScanning(false)
  }

  const handleManualSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!manualCode.trim()) return

    setError(null)
    const data = await lookupOrder(manualCode.trim())

    if (data) {
      setOrderData(data)
      setScanResult(manualCode.trim())
    } else {
      setError("Order not found. Please check the code and try again.")
    }
  }

  const resetScan = () => {
    setScanResult(null)
    setOrderData(null)
    setError(null)
    setManualCode("")
  }

  // QR scanning logic should use a library; no mock auto-detection here
  useEffect(() => {
    // Intentionally left without mock behavior
  }, [isScanning, scanMode])

  // Cleanup camera on unmount
  useEffect(() => {
    return () => stopCamera()
  }, [])

  return (
    <div className="h-full space-y-4 p-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Scanner</h1>
          <p className="text-muted-foreground">Scan QR codes or enter order numbers manually</p>
        </div>

        <Button variant="outline" onClick={resetScan}>
          Reset
        </Button>
      </div>

      {/* Mode Selection */}
      <div className="flex space-x-2">
        <Button
          variant={scanMode === "qr" ? "default" : "outline"}
          onClick={() => setScanMode("qr")}
          className="flex-1"
        >
          <QrCode className="mr-2 h-4 w-4" />
          QR Scanner
        </Button>
        <Button
          variant={scanMode === "manual" ? "default" : "outline"}
          onClick={() => setScanMode("manual")}
          className="flex-1"
        >
          <Hash className="mr-2 h-4 w-4" />
          Manual Entry
        </Button>
      </div>

      {/* Scanner Interface */}
      <div className="grid gap-6 lg:grid-cols-2">
        <Card className="h-fit">
          <CardHeader>
            <CardTitle>{scanMode === "qr" ? "QR Code Scanner" : "Manual Code Entry"}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {scanMode === "qr" ? (
              <div className="space-y-4">
                {/* Camera Interface */}
                <div className="relative">
                  <div className="aspect-video overflow-hidden rounded-lg bg-black">
                    {isScanning ? (
                      <>
                        <video ref={videoRef} autoPlay playsInline className="h-full w-full object-cover" />
                        {/* Scanning overlay */}
                        <div className="absolute inset-0 flex items-center justify-center">
                          <div className="h-48 w-48 animate-pulse rounded-lg border-2 border-dashed border-white">
                            <div className="border-primary absolute top-4 left-4 h-4 w-4 border-t-2 border-l-2"></div>
                            <div className="border-primary absolute top-4 right-4 h-4 w-4 border-t-2 border-r-2"></div>
                            <div className="border-primary absolute bottom-4 left-4 h-4 w-4 border-b-2 border-l-2"></div>
                            <div className="border-primary absolute right-4 bottom-4 h-4 w-4 border-r-2 border-b-2"></div>
                          </div>
                        </div>
                      </>
                    ) : (
                      <div className="flex h-full items-center justify-center text-white">
                        <div className="text-center">
                          <CameraOff className="mx-auto mb-2 h-12 w-12" />
                          <p>Camera inactive</p>
                        </div>
                      </div>
                    )}
                  </div>
                </div>

                {/* Camera Controls */}
                <div className="flex space-x-2">
                  <Button
                    onClick={isScanning ? stopCamera : startCamera}
                    className="flex-1"
                    variant={isScanning ? "destructive" : "default"}
                  >
                    {isScanning ? (
                      <>
                        <CameraOff className="mr-2 h-4 w-4" />
                        Stop Scanner
                      </>
                    ) : (
                      <>
                        <Camera className="mr-2 h-4 w-4" />
                        Start Scanner
                      </>
                    )}
                  </Button>
                </div>
              </div>
            ) : (
              <form onSubmit={handleManualSubmit} className="space-y-4">
                <div>
                  <Label htmlFor="manual-code">Enter Order Number or Table Code</Label>
                  <Input
                    id="manual-code"
                    type="text"
                    placeholder="e.g., BFS-001 or TBL-005"
                    value={manualCode}
                    onChange={(e) => setManualCode(e.target.value)}
                    className="text-lg"
                    autoCapitalize="characters"
                    autoComplete="off"
                  />
                </div>

                <Button type="submit" className="w-full" disabled={!manualCode.trim()}>
                  <Search className="mr-2 h-4 w-4" />
                  Look Up Order
                </Button>
              </form>
            )}

            {/* Error Display */}
            {error && (
              <div className="bg-destructive/10 border-destructive/20 flex items-center rounded-md border p-3">
                <AlertCircle className="text-destructive mr-2 h-4 w-4" />
                <span className="text-destructive text-sm">{error}</span>
              </div>
            )}

            {/* Success Display */}
            {scanResult && !error && (
              <div className="flex items-center rounded-md border border-green-200 bg-green-100 p-3">
                <CheckCircle className="mr-2 h-4 w-4 text-green-600" />
                <span className="text-sm text-green-800">Successfully scanned: {scanResult}</span>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Results Display */}
        <Card className="h-fit">
          <CardHeader>
            <CardTitle>Scan Results</CardTitle>
          </CardHeader>
          <CardContent>
            {orderData ? (
              <div className="space-y-4">
                {"orderNumber" in orderData ? (
                  /* Order Information */
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <h3 className="text-lg font-semibold">{orderData.orderNumber}</h3>
                      <Badge variant="outline" className="bg-green-100 text-green-800">
                        {orderData.status}
                      </Badge>
                    </div>

                    <div className="space-y-2">
                      <div className="flex justify-between text-sm">
                        <span className="text-muted-foreground">Customer:</span>
                        <span className="font-medium">{orderData.customer.name}</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span className="text-muted-foreground">Phone:</span>
                        <span>{orderData.customer.phone}</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span className="text-muted-foreground">Type:</span>
                        <span className="capitalize">{orderData.type}</span>
                      </div>
                    </div>

                    <Separator />

                    <div>
                      <h4 className="mb-2 font-medium">Order Items</h4>
                      {orderData.items.map((item: OrderItem, index: number) => (
                        <div key={index} className="flex justify-between py-1 text-sm">
                          <span>
                            {item.quantity}x {item.name}
                          </span>
                          <span>${(item.price * item.quantity).toFixed(2)}</span>
                        </div>
                      ))}
                    </div>

                    <Separator />

                    <div className="flex justify-between font-semibold">
                      <span>Total:</span>
                      <span>${orderData.total.toFixed(2)}</span>
                    </div>

                    <div className="space-y-2 pt-2">
                      <Button className="w-full">Mark as Picked Up</Button>
                      <Button variant="outline" className="w-full">
                        View Full Order
                      </Button>
                    </div>
                  </div>
                ) : (
                  /* Table Information */
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <h3 className="text-lg font-semibold">{orderData.tableNumber}</h3>
                      <Badge variant="outline" className="bg-blue-100 text-blue-800">
                        {orderData.status}
                      </Badge>
                    </div>

                    <div className="space-y-2">
                      <div className="flex justify-between text-sm">
                        <span className="text-muted-foreground">Capacity:</span>
                        <span>{orderData.capacity} seats</span>
                      </div>
                      {orderData.currentOrder && (
                        <div className="flex justify-between text-sm">
                          <span className="text-muted-foreground">Current Order:</span>
                          <span className="font-medium">{orderData.currentOrder}</span>
                        </div>
                      )}
                    </div>

                    <div className="space-y-2 pt-2">
                      <Button className="w-full">View Table Orders</Button>
                      <Button variant="outline" className="w-full">
                        Clear Table
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <div className="text-muted-foreground py-8 text-center">
                <QrCode className="mx-auto mb-3 h-12 w-12 opacity-50" />
                <p>Scan a QR code or enter a code manually to see order details</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
