"use client"

import { Loader2, Trash2, Upload } from "lucide-react"
import Image from "next/image"
import { useCallback, useRef, useState } from "react"
import { Button } from "@/components/ui/button"

const ACCEPTED_TYPES = "image/jpeg,image/png,image/webp"
const MAX_SIZE = 5 * 1024 * 1024 // 5 MB

type ImageUploadProps = {
  currentImageUrl?: string | null
  onUpload: (file: File) => Promise<string>
  onRemove: () => Promise<void>
  disabled?: boolean
}

export function ImageUpload({ currentImageUrl, onUpload, onRemove, disabled }: ImageUploadProps) {
  const [uploading, setUploading] = useState(false)
  const [removing, setRemoving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const busy = uploading || removing || !!disabled

  const handleFile = useCallback(
    async (file: File) => {
      setError(null)

      if (!file.type.match(/^image\/(jpeg|png|webp)$/)) {
        setError("Nur JPEG, PNG oder WebP erlaubt.")
        return
      }
      if (file.size > MAX_SIZE) {
        setError("Datei darf max. 5 MB gross sein.")
        return
      }

      const objectUrl = URL.createObjectURL(file)
      setPreview(objectUrl)
      setUploading(true)

      try {
        await onUpload(file)
      } catch (e: unknown) {
        setError(e instanceof Error ? e.message : "Upload fehlgeschlagen")
      } finally {
        URL.revokeObjectURL(objectUrl)
        setPreview(null)
        setUploading(false)
      }
    },
    [onUpload]
  )

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) handleFile(file)
    if (inputRef.current) inputRef.current.value = ""
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    const file = e.dataTransfer.files?.[0]
    if (file) handleFile(file)
  }

  const handleRemove = async () => {
    setError(null)
    setRemoving(true)
    try {
      await onRemove()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Löschen fehlgeschlagen")
    } finally {
      setRemoving(false)
    }
  }

  const displayUrl = preview || currentImageUrl

  return (
    <div className="space-y-2">
      <div
        className="relative aspect-video cursor-pointer rounded-[11px] bg-[#cec9c6]"
        onClick={() => !busy && inputRef.current?.click()}
        onDrop={handleDrop}
        onDragOver={(e) => e.preventDefault()}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault()
            if (!busy) inputRef.current?.click()
          }
        }}
        aria-label="Bild hochladen"
      >
        {displayUrl ? (
          <Image
            src={displayUrl}
            alt="Produktbild"
            fill
            sizes="(max-width: 768px) 100vw, (max-width: 1280px) 50vw, 25vw"
            quality={90}
            className="rounded-[11px] object-cover"
            unoptimized={displayUrl.startsWith("blob:") || displayUrl.includes("localhost") || displayUrl.includes("127.0.0.1")}
          />
        ) : (
          <div className="absolute inset-0 flex flex-col items-center justify-center gap-1 text-zinc-500">
            <Upload className="h-6 w-6" />
            <span className="text-xs">Bild hochladen</span>
          </div>
        )}

        {uploading && (
          <div className="absolute inset-0 z-10 flex items-center justify-center rounded-[11px] bg-black/50">
            <Loader2 className="h-6 w-6 animate-spin text-white" />
          </div>
        )}
      </div>

      <input
        ref={inputRef}
        type="file"
        accept={ACCEPTED_TYPES}
        onChange={handleInputChange}
        className="hidden"
        disabled={busy}
      />

      {currentImageUrl && !uploading && (
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="w-full"
          onClick={handleRemove}
          disabled={busy}
        >
          {removing ? <Loader2 className="mr-1 h-3 w-3 animate-spin" /> : <Trash2 className="mr-1 h-3 w-3" />}
          Bild entfernen
        </Button>
      )}

      {error && <p className="text-destructive text-xs">{error}</p>}
    </div>
  )
}
