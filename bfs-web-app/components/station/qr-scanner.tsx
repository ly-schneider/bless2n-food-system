"use client"

import { Camera, Check, Loader2, Package } from "lucide-react"
import { useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Separator } from "@/components/ui/separator"
import { StationAuthService } from "@/lib/station-auth"
import { RedemptionItem } from "@/types"

interface QRScannerProps {
  stationId: string
}

type RedemptionOrder = {
  id: string
  total: number
  status: string
  items?: { productId: string; quantity: number; name: string; price: number }[]
}

export function QRScanner({ stationId: _stationId }: QRScannerProps) {
  const [isScanning, setIsScanning] = useState(false)
  const [manualOrderId, setManualOrderId] = useState("")
  const [currentOrder, setCurrentOrder] = useState<RedemptionOrder | null>(null)
  const [selectedItems, setSelectedItems] = useState<RedemptionItem[]>([])
  const [isProcessing, setIsProcessing] = useState(false)
  const [redemptionComplete, setRedemptionComplete] = useState(false)
  const videoRef = useRef<HTMLVideoElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)

  const startCamera = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: {
          facingMode: "environment", // Use back camera on mobile
          width: { ideal: 1280 },
          height: { ideal: 720 },
        },
      })

      if (videoRef.current) {
        videoRef.current.srcObject = stream
        setIsScanning(true)
      }
    } catch (error) {
      console.error("Failed to start camera:", error)
      alert("Failed to access camera. Please allow camera permissions.")
    }
  }

  const stopCamera = () => {
    if (videoRef.current?.srcObject) {
      const stream = videoRef.current.srcObject as MediaStream
      stream.getTracks().forEach((track) => track.stop())
      videoRef.current.srcObject = null
    }
    setIsScanning(false)
  }

  // const handleQRDetected = async (orderId: string) => {
  //   stopCamera()
  //   await loadOrder(orderId)
  // }

  const loadOrder = async (orderId: string) => {
    try {
      setIsProcessing(true)
      const order = (await StationAuthService.getOrderForRedemption(orderId)) as RedemptionOrder
      setCurrentOrder(order as RedemptionOrder)
      // Initialize with all items selected for redemption
      setSelectedItems(
        order.items?.map((item: { productId: string; quantity: number }) => ({
          productId: item.productId,
          quantity: item.quantity,
        })) || []
      )
      setManualOrderId("")
    } catch (error) {
      console.error("Failed to load order:", error)
      alert("Order not found or not eligible for redemption.")
      setCurrentOrder(null)
    } finally {
      setIsProcessing(false)
    }
  }

  const handleManualEntry = () => {
    if (manualOrderId.trim()) {
      loadOrder(manualOrderId.trim())
    }
  }

  const toggleItemSelection = (productId: string, _quantity: number) => {
    setSelectedItems((prev) => {
      const existing = prev.find((item) => item.productId === productId)
      if (existing) {
        // Remove item or reduce quantity
        if (existing.quantity > 1) {
          return prev
            .map((item) =>
              item.productId === productId ? { ...item, quantity: Math.max(0, item.quantity - 1) } : item
            )
            .filter((item) => item.quantity > 0)
        } else {
          return prev.filter((item) => item.productId !== productId)
        }
      } else {
        // Add item
        return [...prev, { productId, quantity: 1 }]
      }
    })
  }

  const processRedemption = async () => {
    if (!currentOrder || selectedItems.length === 0) return

    try {
      setIsProcessing(true)
      await StationAuthService.redeemOrder(currentOrder.id, selectedItems)
      setRedemptionComplete(true)

      // Reset after 3 seconds
      setTimeout(() => {
        setCurrentOrder(null)
        setSelectedItems([])
        setRedemptionComplete(false)
      }, 3000)
    } catch (error) {
      console.error("Failed to process redemption:", error)
      alert("Failed to process redemption. Please try again.")
    } finally {
      setIsProcessing(false)
    }
  }

  const resetScanner = () => {
    setCurrentOrder(null)
    setSelectedItems([])
    setRedemptionComplete(false)
    setManualOrderId("")
  }

  if (redemptionComplete) {
    return (
      <Card className="w-full">
        <CardContent className="py-8 text-center">
          <Check className="mx-auto mb-4 h-16 w-16 text-green-600" />
          <h3 className="mb-2 text-2xl font-bold text-green-600">Redemption Complete!</h3>
          <p className="text-gray-600">Items have been successfully redeemed.</p>
        </CardContent>
      </Card>
    )
  }

  if (currentOrder) {
    return (
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Order #{currentOrder.id.slice(-8)}</CardTitle>
                <CardDescription>
                  Total: ${currentOrder.total.toFixed(2)} • Status: {currentOrder.status}
                </CardDescription>
              </div>
              <Button variant="outline" size="sm" onClick={resetScanner}>
                Cancel
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <h4 className="font-medium">Select items to redeem:</h4>

              {currentOrder.items?.map((item: { productId: string; quantity: number; name: string; price: number }) => {
                const selectedItem = selectedItems.find((si) => si.productId === item.productId)
                const selectedQuantity = selectedItem?.quantity || 0

                return (
                  <div key={item.productId} className="flex items-center justify-between rounded-lg border p-3">
                    <div className="flex items-center gap-3">
                      <Package className="h-5 w-5 text-gray-400" />
                      <div>
                        <p className="font-medium">{item.name}</p>
                        <p className="text-sm text-gray-600">
                          Ordered: {item.quantity} • ${item.price.toFixed(2)} each
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => toggleItemSelection(item.productId, item.quantity)}
                        className={selectedQuantity > 0 ? "border-blue-200 bg-blue-50" : ""}
                      >
                        {selectedQuantity > 0 ? (
                          <span className="flex items-center gap-1">
                            <Check className="h-3 w-3" />
                            {selectedQuantity}
                          </span>
                        ) : (
                          "Select"
                        )}
                      </Button>
                    </div>
                  </div>
                )
              })}

              <Separator />

              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Items selected: {selectedItems.length}</p>
                  <p className="text-sm text-gray-600">
                    Total quantity: {selectedItems.reduce((sum, item) => sum + item.quantity, 0)}
                  </p>
                </div>

                <Button
                  onClick={processRedemption}
                  disabled={selectedItems.length === 0 || isProcessing}
                  className="bg-green-600 hover:bg-green-700"
                >
                  {isProcessing ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Processing...
                    </>
                  ) : (
                    "Complete Redemption"
                  )}
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* QR Scanner */}
      <Card>
        <CardHeader>
          <CardTitle>QR Code Scanner</CardTitle>
          <CardDescription>Scan customer order QR codes or enter order ID manually</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {!isScanning ? (
            <div className="space-y-4">
              <div className="flex aspect-video items-center justify-center rounded-lg border-2 border-dashed border-gray-300 bg-gray-100">
                <div className="text-center">
                  <Camera className="mx-auto mb-3 h-12 w-12 text-gray-400" />
                  <p className="mb-4 text-gray-600">Ready to scan QR codes</p>
                  <Button onClick={startCamera}>Start Camera</Button>
                </div>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="relative aspect-video overflow-hidden rounded-lg bg-black">
                <video ref={videoRef} autoPlay playsInline className="h-full w-full object-cover" />
                <canvas ref={canvasRef} className="absolute inset-0 h-full w-full" style={{ display: "none" }} />

                {/* QR Code overlay */}
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className="relative h-64 w-64 rounded-lg border-2 border-white">
                    <div className="absolute top-0 left-0 h-8 w-8 rounded-tl-lg border-t-4 border-l-4 border-blue-500"></div>
                    <div className="absolute top-0 right-0 h-8 w-8 rounded-tr-lg border-t-4 border-r-4 border-blue-500"></div>
                    <div className="absolute bottom-0 left-0 h-8 w-8 rounded-bl-lg border-b-4 border-l-4 border-blue-500"></div>
                    <div className="absolute right-0 bottom-0 h-8 w-8 rounded-br-lg border-r-4 border-b-4 border-blue-500"></div>
                  </div>
                </div>
              </div>

              <div className="flex justify-center">
                <Button variant="outline" onClick={stopCamera}>
                  Stop Camera
                </Button>
              </div>
            </div>
          )}

          <Separator />

          {/* Manual Entry */}
          <div className="space-y-3">
            <Label htmlFor="manual-order">Or enter Order ID manually:</Label>
            <div className="flex gap-2">
              <Input
                id="manual-order"
                placeholder="Enter order ID"
                value={manualOrderId}
                onChange={(e) => setManualOrderId(e.target.value)}
                onKeyPress={(e) => e.key === "Enter" && handleManualEntry()}
              />
              <Button onClick={handleManualEntry} disabled={!manualOrderId.trim() || isProcessing}>
                {isProcessing ? <Loader2 className="h-4 w-4 animate-spin" /> : "Load Order"}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Redemptions</CardTitle>
          <CardDescription>Latest orders processed at this station</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="py-8 text-center text-gray-500">
            <Package className="mx-auto mb-2 h-8 w-8" />
            <p>No recent redemptions</p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
