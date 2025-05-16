'use client'

import { useState, useEffect } from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { PlusCircle, Pencil, Trash2, Eye } from 'lucide-react'
import { toast } from "sonner"
import { useSupabase } from '@/hooks/useSupabase'

interface Article {
  id: string
  name: string
  price: number
  description?: string
  category?: string
  // Add other fields as needed
}

export function ArticlesSection() {
  const { supabase } = useSupabase()
  const [articles, setArticles] = useState<Article[]>([])
  const [selectedArticle, setSelectedArticle] = useState<Article | null>(null)
  const [isViewDialogOpen, setIsViewDialogOpen] = useState(false)
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false)
  const [isNewDialogOpen, setIsNewDialogOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    fetchArticles()
  }, [])

  const fetchArticles = async () => {
    try {
      setIsLoading(true)
      const { data, error } = await supabase.functions.invoke<Article[]>('get_item', {
        method: 'GET',
        body: {}, // Empty body for GET request
      })

      if (error) {
        console.error('Error fetching articles:', error)
        toast.error("Failed to fetch articles")
        throw error
      }

      if (!data) {
        throw new Error('No data received from API')
      }

      setArticles(data)
    } catch (error) {
      console.error('Error fetching articles:', error)
      toast.error("Error loading articles")
    } finally {
      setIsLoading(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this article?')) return
    
    // For now, just log the deletion
    console.log('Delete article:', id)
    toast.success("Delete functionality will be implemented soon")
  }

  const ArticleForm = ({ article, onSubmit }: { article?: Article, onSubmit: (data: Partial<Article>) => void }) => (
    <form onSubmit={(e) => {
      e.preventDefault()
      const formData = new FormData(e.currentTarget)
      onSubmit({
        name: formData.get('name') as string,
        price: parseFloat(formData.get('price') as string),
        description: formData.get('description') as string,
        category: formData.get('category') as string,
      })
    }} className="space-y-4">
      <div>
        <Label htmlFor="name">Name</Label>
        <Input 
          id="name" 
          name="name" 
          defaultValue={article?.name} 
          className="mt-1.5"
          placeholder="Artikelname eingeben"
          required 
        />
      </div>
      <div>
        <Label htmlFor="price">Preis (CHF)</Label>
        <Input 
          id="price" 
          name="price" 
          type="number" 
          step="0.01" 
          defaultValue={article?.price}
          className="mt-1.5 font-mono"
          placeholder="0.00"
          required 
        />
      </div>
      <div>
        <Label htmlFor="description">Beschreibung</Label>
        <Input 
          id="description" 
          name="description" 
          defaultValue={article?.description}
          className="mt-1.5"
          placeholder="Artikelbeschreibung eingeben"
        />
      </div>
      <div>
        <Label htmlFor="category">Kategorie</Label>
        <Input 
          id="category" 
          name="category" 
          defaultValue={article?.category}
          className="mt-1.5"
          placeholder="Kategorie eingeben"
        />
      </div>
      <div className="pt-4">
        <Button type="submit" className="w-full">
          {article ? 'Speichern' : 'Erstellen'}
        </Button>
      </div>
    </form>
  )

  return (
    <div className="h-full">
      <div className="p-8">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h2 className="text-3xl font-bold text-gray-900">Artikel Verwaltung</h2>
            <p className="mt-2 text-sm text-muted-foreground">
              Verwalten Sie hier Ihre Artikel und Produkte
            </p>
          </div>
          <Dialog open={isNewDialogOpen} onOpenChange={setIsNewDialogOpen}>
            <DialogTrigger asChild>
              <Button className="bg-primary hover:bg-primary/90" size="lg">
                <PlusCircle className="mr-2 h-5 w-5" />
                Neuer Artikel
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[500px]">
              <DialogHeader>
                <DialogTitle className="text-xl">Neuer Artikel</DialogTitle>
              </DialogHeader>
              <ArticleForm onSubmit={async (data) => {
                console.log('Create article:', data)
                toast.success("Artikel wird bald erstellt")
                setIsNewDialogOpen(false)
              }} />
            </DialogContent>
          </Dialog>
        </div>

        <div className="bg-white shadow-sm rounded-lg border border-gray-200">
          <div className="p-4 border-b border-gray-200">
            <h3 className="font-semibold text-gray-900">Artikelliste</h3>
          </div>
          <Table>
            <TableHeader>
              <TableRow className="bg-gray-50 hover:bg-gray-50">
                <TableHead className="w-[40%] py-4 font-medium">Name</TableHead>
                <TableHead className="w-[20%] font-medium">Preis</TableHead>
                <TableHead className="w-[25%] font-medium">Kategorie</TableHead>
                <TableHead className="w-[15%] text-right font-medium">Aktionen</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center py-12">
                    <div className="flex flex-col items-center gap-3">
                      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                      <p className="text-sm text-muted-foreground">Artikel werden geladen...</p>
                    </div>
                  </TableCell>
                </TableRow>
              ) : articles.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center py-12">
                    <div className="flex flex-col items-center gap-2">
                      <p className="text-sm text-muted-foreground">Keine Artikel gefunden</p>
                      <Button 
                        variant="link" 
                        onClick={() => setIsNewDialogOpen(true)}
                        className="text-primary hover:text-primary/90"
                      >
                        Erstellen Sie Ihren ersten Artikel
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ) : (
                articles.map((article) => (
                  <TableRow key={article.id} className="hover:bg-gray-50">
                    <TableCell className="py-4 font-medium">{article.name}</TableCell>
                    <TableCell className="font-mono">{article.price.toFixed(2)} CHF</TableCell>
                    <TableCell>{article.category || '-'}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="hover:bg-gray-100"
                          onClick={() => {
                            setSelectedArticle(article)
                            setIsViewDialogOpen(true)
                          }}
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="hover:bg-gray-100"
                          onClick={() => {
                            setSelectedArticle(article)
                            setIsEditDialogOpen(true)
                          }}
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="hover:bg-gray-100 hover:text-red-600"
                          onClick={() => handleDelete(article.id)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      {/* View Dialog */}
      <Dialog open={isViewDialogOpen} onOpenChange={setIsViewDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle className="text-xl">Artikel Details</DialogTitle>
          </DialogHeader>
          {selectedArticle && (
            <div className="space-y-6">
              <div>
                <Label className="text-sm font-medium text-gray-500">Name</Label>
                <p className="mt-1 text-lg font-medium">{selectedArticle.name}</p>
              </div>
              <div>
                <Label className="text-sm font-medium text-gray-500">Preis</Label>
                <p className="mt-1 text-lg font-medium font-mono">{selectedArticle.price.toFixed(2)} CHF</p>
              </div>
              <div>
                <Label className="text-sm font-medium text-gray-500">Beschreibung</Label>
                <p className="mt-1 text-lg">{selectedArticle.description || '-'}</p>
              </div>
              <div>
                <Label className="text-sm font-medium text-gray-500">Kategorie</Label>
                <p className="mt-1 text-lg">{selectedArticle.category || '-'}</p>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle className="text-xl">Artikel Bearbeiten</DialogTitle>
          </DialogHeader>
          {selectedArticle && (
            <ArticleForm 
              article={selectedArticle} 
              onSubmit={async (data) => {
                console.log('Update article:', { id: selectedArticle.id, ...data })
                toast.success("Artikel wird bald aktualisiert")
                setIsEditDialogOpen(false)
              }} 
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}
