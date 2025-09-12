"use client"

import {
  Calendar,
  Loader2,
  Mail,
  MoreHorizontal,
  Search,
  Shield,
  Trash2,
  UserCheck,
  UserPlus,
  Users,
  UserX,
} from "lucide-react"
import { useEffect, useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Textarea } from "@/components/ui/textarea"
import AdminAPI, { AdminCustomer, BanCustomerRequest, InviteAdminRequest } from "@/lib/api/admin"

export default function CustomersPage() {
  const [customers, setCustomers] = useState<AdminCustomer[]>([])
  const [filteredCustomers, setFilteredCustomers] = useState<AdminCustomer[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState("")
  const [showBanDialog, setBanDialog] = useState<{ customer: AdminCustomer; action: "ban" | "unban" } | null>(null)
  const [showDeleteDialog, setDeleteDialog] = useState<AdminCustomer | null>(null)
  const [showInviteDialog, setShowInviteDialog] = useState(false)
  const [banReason, setBanReason] = useState("")

  useEffect(() => {
    loadCustomers()
  }, [])

  useEffect(() => {
    const filtered = customers.filter(
      (customer) =>
        customer.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        customer.email.toLowerCase().includes(searchQuery.toLowerCase())
    )
    setFilteredCustomers(filtered)
  }, [customers, searchQuery])

  const loadCustomers = async () => {
    try {
      setIsLoading(true)
      const response = await AdminAPI.listCustomers({ limit: 100 })
      setCustomers(response.customers)
    } catch (error) {
      console.error("Failed to load customers:", error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleBanCustomer = async (customerId: string, banned: boolean, reason?: string) => {
    try {
      const request: BanCustomerRequest = { banned, reason }
      await AdminAPI.banCustomer(customerId, request)

      // Update local state
      setCustomers((prev) =>
        prev.map((customer) => (customer.id === customerId ? { ...customer, isActive: !banned } : customer))
      )

      setBanDialog(null)
      setBanReason("")
    } catch (error) {
      console.error("Failed to update customer status:", error)
      alert("Failed to update customer status. Please try again.")
    }
  }

  const handleDeleteCustomer = async (customerId: string) => {
    try {
      await AdminAPI.deleteCustomer(customerId)
      setCustomers((prev) => prev.filter((customer) => customer.id !== customerId))
      setDeleteDialog(null)
    } catch (error) {
      console.error("Failed to delete customer:", error)
      alert("Failed to delete customer. Please try again.")
    }
  }

  const handleInviteAdmin = async (inviteData: InviteAdminRequest) => {
    try {
      await AdminAPI.inviteAdmin(inviteData)
      setShowInviteDialog(false)
      alert("Admin invitation sent successfully!")
    } catch (error) {
      console.error("Failed to invite admin:", error)
      alert("Failed to send admin invitation. Please try again.")
    }
  }

  const getCustomerStats = () => {
    const total = customers.length
    const active = customers.filter((c) => c.isActive).length
    const verified = customers.filter((c) => c.isEmailVerified).length

    return { total, active, verified }
  }

  const stats = getCustomerStats()

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mx-auto max-w-7xl">
        <div className="mb-8 flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Customer Management</h1>
            <p className="text-gray-600">Manage customer accounts and permissions</p>
          </div>

          <Dialog open={showInviteDialog} onOpenChange={setShowInviteDialog}>
            <DialogTrigger asChild>
              <Button>
                <UserPlus className="mr-2 h-4 w-4" />
                Invite Admin
              </Button>
            </DialogTrigger>
            <DialogContent>
              <AdminInviteForm onSubmit={handleInviteAdmin} onCancel={() => setShowInviteDialog(false)} />
            </DialogContent>
          </Dialog>
        </div>

        {/* Stats Cards */}
        <div className="mb-8 grid gap-6 md:grid-cols-3">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Customers</CardTitle>
              <Users className="text-muted-foreground h-4 w-4" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total}</div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Active Customers</CardTitle>
              <UserCheck className="h-4 w-4 text-green-600" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-600">{stats.active}</div>
              <p className="text-muted-foreground text-xs">
                {stats.total > 0 ? Math.round((stats.active / stats.total) * 100) : 0}% of total
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Verified Emails</CardTitle>
              <Mail className="h-4 w-4 text-blue-600" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-blue-600">{stats.verified}</div>
              <p className="text-muted-foreground text-xs">
                {stats.total > 0 ? Math.round((stats.verified / stats.total) * 100) : 0}% verified
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Search and Filters */}
        <Card className="mb-6">
          <CardHeader>
            <div className="flex items-center gap-4">
              <div className="relative max-w-sm flex-1">
                <Search className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 transform text-gray-400" />
                <Input
                  placeholder="Search customers..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>
          </CardHeader>
        </Card>

        {/* Customer Table */}
        <Card>
          <CardHeader>
            <CardTitle>Customers</CardTitle>
            <CardDescription>
              {filteredCustomers.length} {filteredCustomers.length === 1 ? "customer" : "customers"} found
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Customer</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Joined</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredCustomers.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="py-8 text-center text-gray-500">
                      {searchQuery ? "No customers found matching your search." : "No customers yet."}
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredCustomers.map((customer) => (
                    <TableRow key={customer.id}>
                      <TableCell>
                        <div>
                          <div className="font-medium">{customer.name}</div>
                          <div className="text-sm text-gray-500">ID: {customer.id.slice(-8)}</div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {customer.email}
                          {customer.isEmailVerified && (
                            <Badge variant="outline" className="text-xs">
                              Verified
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={customer.isActive ? "default" : "destructive"}>
                          {customer.isActive ? "Active" : "Banned"}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Calendar className="h-4 w-4 text-gray-400" />
                          {new Date(customer.createdAt).toLocaleDateString()}
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" className="h-8 w-8 p-0">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={() =>
                                setBanDialog({
                                  customer,
                                  action: customer.isActive ? "ban" : "unban",
                                })
                              }
                            >
                              {customer.isActive ? (
                                <>
                                  <UserX className="mr-2 h-4 w-4" />
                                  Ban Customer
                                </>
                              ) : (
                                <>
                                  <UserCheck className="mr-2 h-4 w-4" />
                                  Unban Customer
                                </>
                              )}
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem onClick={() => setDeleteDialog(customer)} className="text-red-600">
                              <Trash2 className="mr-2 h-4 w-4" />
                              Delete Customer
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        {/* Ban/Unban Dialog */}
        <Dialog open={!!showBanDialog} onOpenChange={(open) => !open && setBanDialog(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{showBanDialog?.action === "ban" ? "Ban Customer" : "Unban Customer"}</DialogTitle>
              <DialogDescription>
                {showBanDialog?.action === "ban"
                  ? "This will prevent the customer from accessing their account and placing orders."
                  : "This will restore the customer's access to their account."}
              </DialogDescription>
            </DialogHeader>

            {showBanDialog?.action === "ban" && (
              <div className="space-y-2">
                <Label htmlFor="ban-reason">Reason for banning (optional)</Label>
                <Textarea
                  id="ban-reason"
                  value={banReason}
                  onChange={(e) => setBanReason(e.target.value)}
                  placeholder="Enter reason for banning this customer..."
                />
              </div>
            )}

            <DialogFooter>
              <Button variant="outline" onClick={() => setBanDialog(null)}>
                Cancel
              </Button>
              <Button
                variant={showBanDialog?.action === "ban" ? "destructive" : "default"}
                onClick={() => {
                  if (showBanDialog) {
                    handleBanCustomer(showBanDialog.customer.id, showBanDialog.action === "ban", banReason || undefined)
                  }
                }}
              >
                {showBanDialog?.action === "ban" ? "Ban Customer" : "Unban Customer"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Delete Dialog */}
        <Dialog open={!!showDeleteDialog} onOpenChange={(open) => !open && setDeleteDialog(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete Customer</DialogTitle>
              <DialogDescription>
                Are you sure you want to permanently delete {showDeleteDialog?.name}? This action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDeleteDialog(null)}>
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={() => {
                  if (showDeleteDialog) {
                    handleDeleteCustomer(showDeleteDialog.id)
                  }
                }}
              >
                Delete Customer
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  )
}

interface AdminInviteFormProps {
  onSubmit: (data: InviteAdminRequest) => void
  onCancel: () => void
}

function AdminInviteForm({ onSubmit, onCancel }: AdminInviteFormProps) {
  const [formData, setFormData] = useState({
    email: "",
    name: "",
  })
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    try {
      await onSubmit(formData)
      setFormData({ email: "", name: "" })
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5" />
          Invite New Admin
        </DialogTitle>
        <DialogDescription>Send an invitation to create a new admin account.</DialogDescription>
      </DialogHeader>

      <form onSubmit={handleSubmit}>
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="admin-email">Email Address *</Label>
            <Input
              id="admin-email"
              type="email"
              required
              value={formData.email}
              onChange={(e) => setFormData((prev) => ({ ...prev, email: e.target.value }))}
              placeholder="admin@example.com"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="admin-name">Full Name *</Label>
            <Input
              id="admin-name"
              required
              value={formData.name}
              onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
              placeholder="Admin Name"
            />
          </div>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Sending...
              </>
            ) : (
              "Send Invitation"
            )}
          </Button>
        </DialogFooter>
      </form>
    </>
  )
}
