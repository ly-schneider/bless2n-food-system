"use client"

import { CheckCircle2, Clock, Loader2, LogOut, Store, XCircle } from "lucide-react"
import { useEffect, useState } from "react"
import { QRScanner } from "@/components/station/qr-scanner"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { StationAuthService } from "@/lib/station-auth"
import type { StationRequestForm as StationRequestFormData, StationSession } from "@/types"

export default function StationPage() {
  const [station, setStation] = useState<StationSession | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [showRequestForm, setShowRequestForm] = useState(false)
  const [requestSubmitted, setRequestSubmitted] = useState(false)

  useEffect(() => {
    checkStationStatus()
  }, [])

  const checkStationStatus = async () => {
    try {
      const currentStation = await StationAuthService.getCurrentStation()
      setStation(currentStation)
    } catch (error) {
      console.error("Failed to check station status:", error)
    } finally {
      setIsLoading(false)
    }
  }

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    )
  }

  // Station is authenticated and approved
  if (station?.isAuthenticated && station.status === "approved") {
    return <StationDashboard station={station} />
  }

  // Station has pending request
  if (station && station.status === "pending") {
    return <PendingApproval station={station} />
  }

  // Station request was rejected
  if (station && station.status === "rejected") {
    return <RequestRejected onRequestAgain={() => setShowRequestForm(true)} />
  }

  // No station session - show request form or landing
  if (showRequestForm || requestSubmitted) {
    return <StationRequestForm onSuccess={() => setRequestSubmitted(true)} isSubmitted={requestSubmitted} />
  }

  return <StationLanding onRequestAccess={() => setShowRequestForm(true)} />
}

function StationLanding({ onRequestAccess }: { onRequestAccess: () => void }) {
  return (
    <div className="min-h-screen bg-gradient-to-b from-blue-50 to-white">
      <div className="container mx-auto px-4 py-16">
        <div className="mx-auto max-w-4xl text-center">
          <Store className="mx-auto mb-6 h-16 w-16 text-blue-600" />
          <h1 className="mb-6 text-4xl font-bold text-gray-900">Partner Station Portal</h1>
          <p className="mx-auto mb-8 max-w-2xl text-xl text-gray-600">
            Join the Bless2n Food Network as a partner station. Serve fresh, quality meals to your community while being
            part of our sustainable food ecosystem.
          </p>

          <div className="mb-12 grid gap-8 md:grid-cols-3">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Store className="h-5 w-5" />
                  Easy Setup
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-gray-600">
                  Quick onboarding process with dedicated support to get your station running.
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <CheckCircle2 className="h-5 w-5" />
                  Quality Products
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-gray-600">Access to our curated selection of fresh, sustainable food products.</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Clock className="h-5 w-5" />
                  Flexible Operations
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-gray-600">Operate on your schedule with our flexible station management system.</p>
              </CardContent>
            </Card>
          </div>

          <Button onClick={onRequestAccess} size="lg" className="bg-blue-600 hover:bg-blue-700">
            Request Station Access
          </Button>
        </div>
      </div>
    </div>
  )
}

function StationRequestForm({ onSuccess, isSubmitted }: { onSuccess: () => void; isSubmitted: boolean }) {
  const [formData, setFormData] = useState<StationRequestFormData>({
    businessName: "",
    contactEmail: "",
    contactName: "",
    location: "",
    description: "",
    businessType: "",
    operatingHours: "",
  })
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)

    try {
      await StationAuthService.requestStationAccess(formData)
      onSuccess()
    } catch (error) {
      console.error("Failed to submit station request:", error)
      alert("Failed to submit request. Please try again.")
    } finally {
      setIsLoading(false)
    }
  }

  if (isSubmitted) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <CheckCircle2 className="mx-auto mb-4 h-12 w-12 text-green-600" />
            <CardTitle>Request Submitted!</CardTitle>
            <CardDescription>
              We've received your station access request. Our team will review it and contact you within 48 hours.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-center text-sm text-gray-600">
              Check your email for updates on your application status.
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 py-12">
      <div className="container mx-auto px-4">
        <div className="mx-auto max-w-2xl">
          <Card>
            <CardHeader>
              <CardTitle>Request Station Access</CardTitle>
              <CardDescription>Fill out the form below to request access as a partner station.</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="space-y-6">
                <div className="grid gap-4 md:grid-cols-2">
                  <div className="space-y-2">
                    <Label htmlFor="businessName">Business Name *</Label>
                    <Input
                      id="businessName"
                      required
                      value={formData.businessName}
                      onChange={(e) => setFormData((prev) => ({ ...prev, businessName: e.target.value }))}
                      placeholder="Your business name"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="contactName">Contact Person *</Label>
                    <Input
                      id="contactName"
                      required
                      value={formData.contactName}
                      onChange={(e) => setFormData((prev) => ({ ...prev, contactName: e.target.value }))}
                      placeholder="Your full name"
                    />
                  </div>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="contactEmail">Contact Email *</Label>
                  <Input
                    id="contactEmail"
                    type="email"
                    required
                    value={formData.contactEmail}
                    onChange={(e) => setFormData((prev) => ({ ...prev, contactEmail: e.target.value }))}
                    placeholder="your@email.com"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="location">Location *</Label>
                  <Input
                    id="location"
                    required
                    value={formData.location}
                    onChange={(e) => setFormData((prev) => ({ ...prev, location: e.target.value }))}
                    placeholder="City, Country or full address"
                  />
                </div>

                <div className="grid gap-4 md:grid-cols-2">
                  <div className="space-y-2">
                    <Label htmlFor="businessType">Business Type</Label>
                    <Input
                      id="businessType"
                      value={formData.businessType}
                      onChange={(e) => setFormData((prev) => ({ ...prev, businessType: e.target.value }))}
                      placeholder="e.g., Restaurant, Cafe, Market"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="operatingHours">Operating Hours</Label>
                    <Input
                      id="operatingHours"
                      value={formData.operatingHours}
                      onChange={(e) => setFormData((prev) => ({ ...prev, operatingHours: e.target.value }))}
                      placeholder="e.g., Mon-Fri 9AM-5PM"
                    />
                  </div>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="description">Description</Label>
                  <Textarea
                    id="description"
                    value={formData.description}
                    onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
                    placeholder="Tell us about your business and why you'd like to become a partner station"
                    rows={4}
                  />
                </div>

                <Button type="submit" className="w-full" disabled={isLoading}>
                  {isLoading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Submitting...
                    </>
                  ) : (
                    "Submit Request"
                  )}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

function PendingApproval({ station }: { station: StationSession }) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <Clock className="mx-auto mb-4 h-12 w-12 text-yellow-600" />
          <CardTitle>Application Under Review</CardTitle>
          <CardDescription>
            Your station access request for "{station.stationName}" is currently being reviewed by our team.
          </CardDescription>
        </CardHeader>
        <CardContent className="text-center">
          <p className="mb-4 text-sm text-gray-600">
            We'll notify you via email once your application has been processed.
          </p>
          <div className="rounded-lg bg-yellow-50 p-3 text-sm text-yellow-800">Expected review time: 24-48 hours</div>
        </CardContent>
      </Card>
    </div>
  )
}

function RequestRejected({ onRequestAgain }: { onRequestAgain: () => void }) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <XCircle className="mx-auto mb-4 h-12 w-12 text-red-600" />
          <CardTitle>Application Not Approved</CardTitle>
          <CardDescription>Unfortunately, your station access request was not approved at this time.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 text-center">
          <p className="text-sm text-gray-600">
            This may be due to location coverage, capacity constraints, or other factors. You're welcome to submit a new
            application with updated information.
          </p>
          <Button onClick={onRequestAgain} variant="outline">
            Submit New Request
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}

function StationDashboard({ station }: { station: StationSession }) {
  const handleLogout = async () => {
    try {
      await StationAuthService.logoutStation()
      window.location.reload() // Refresh to show logged out state
    } catch (error) {
      console.error("Logout failed:", error)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <div className="mx-auto max-w-6xl">
          <div className="mb-8 flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Station Dashboard</h1>
              <p className="text-gray-600">
                {station.stationName} • {station.location}
              </p>
            </div>
            <Button variant="outline" onClick={handleLogout}>
              <LogOut className="mr-2 h-4 w-4" />
              Logout
            </Button>
          </div>

          <div className="grid gap-8 lg:grid-cols-3">
            <div className="lg:col-span-2">
              <QRScanner stationId={station.stationId} />
            </div>

            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Station Status</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex justify-between">
                      <span className="text-sm font-medium">Status</span>
                      <span className="text-sm text-green-600 capitalize">{station.status}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm font-medium">Location</span>
                      <span className="text-sm text-gray-600">{station.location}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm font-medium">Station ID</span>
                      <span className="font-mono text-sm text-gray-600">{station.stationId.slice(-8)}</span>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Today's Activity</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex justify-between">
                      <span className="text-sm font-medium">Orders Redeemed</span>
                      <span className="text-sm text-gray-600">0</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm font-medium">Items Dispensed</span>
                      <span className="text-sm text-gray-600">0</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm font-medium">Revenue</span>
                      <span className="text-sm text-gray-600">$0.00</span>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Station Products</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="mb-3 text-sm text-gray-600">Products available at this station</p>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      /* TODO: View station products */
                    }}
                  >
                    View Catalog
                  </Button>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Support</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="mb-3 text-sm text-gray-600">Need help with your station?</p>
                  <Button variant="outline" size="sm">
                    Contact Support
                  </Button>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
