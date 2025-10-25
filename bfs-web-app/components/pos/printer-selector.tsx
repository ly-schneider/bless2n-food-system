"use client"

import { Bluetooth, BluetoothSearching, Check, Loader2, RefreshCw, WifiOff } from "lucide-react"
import { useEffect, useMemo, useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"

type Bridge = {
  listPrinters?: () => string
  selectPrinter?: (address?: string | null) => void
  getSelectedPrinter?: () => string
  startDiscovery?: () => boolean
  stopDiscovery?: () => void
  pair?: (address?: string | null) => boolean
}

type Printer = { name: string; address: string; protocol?: string; bonded?: boolean }

export function PrinterSelector() {
  const [open, setOpen] = useState(false)
  const [selected, setSelected] = useState<string>("")
  const [paired, setPaired] = useState<Record<string, Printer>>({})
  const [found, setFound] = useState<Record<string, Printer>>({})
  const [scanning, setScanning] = useState(false)
  const [busy, setBusy] = useState(false)
  const [hasBridge, setHasBridge] = useState(false)
  const listenersBound = useRef(false)

  function bridge(): Bridge | undefined {
    try {
      return (globalThis as any)?.PosBridge as Bridge
    } catch {
      return undefined
    }
  }

  async function refreshPaired() {
    const b = bridge()
    if (!b?.listPrinters) return
    try {
      const raw = b.listPrinters()
      const arr = JSON.parse(raw || "[]") as Array<Printer>
      const next: Record<string, Printer> = {}
      for (const p of arr) if (p.address) next[p.address] = { ...p, bonded: true }
      setPaired(next)
    } catch {}
  }

  function refreshSelected() {
    const b = bridge()
    try {
      const cur = b?.getSelectedPrinter?.() || ""
      setSelected(cur)
    } catch {}
  }

  useEffect(() => {
    const b = bridge()
    setHasBridge(!!b)
    refreshSelected()
    refreshPaired()

    if (!listenersBound.current) {
      const onFound = (ev: Event) => {
        try {
          const d = (ev as CustomEvent<Printer>).detail
          if (!d?.address) return
          setFound((prev) => ({ ...prev, [d.address]: { ...d } }))
        } catch {}
      }
      const onDone = () => setScanning(false)
      const onBond = (ev: Event) => {
        try {
          const d = (ev as CustomEvent<{ address: string; state: string }>).detail
          if (!d?.address) return
          const bonded = d.state === "bonded"
          setPaired((prev) => {
            const cur = { ...prev }
            const name = cur[d.address]?.name || found[d.address]?.name || d.address
            const proto = cur[d.address]?.protocol || found[d.address]?.protocol
            if (bonded) cur[d.address] = { name, address: d.address, bonded: true, protocol: proto }
            else delete cur[d.address]
            return cur
          })
          setFound((prev) => {
            const cur = { ...prev }
            const item = cur[d.address]
            if (item) cur[d.address] = { ...item, bonded }
            return cur
          })
        } catch {}
      }
      window.addEventListener("bfs:printer:found", onFound as EventListener)
      window.addEventListener("bfs:printer:discovery:finished", onDone)
      window.addEventListener("bfs:printer:bond:state", onBond as EventListener)
      listenersBound.current = true
      return () => {
        window.removeEventListener("bfs:printer:found", onFound as EventListener)
        window.removeEventListener("bfs:printer:discovery:finished", onDone)
        window.removeEventListener("bfs:printer:bond:state", onBond as EventListener)
        listenersBound.current = false
      }
    }
  }, [])

  const selectedName = useMemo(() => {
    if (!selected) return null
    return paired[selected]?.name || found[selected]?.name || selected
  }, [selected, paired, found])

  function startScan() {
    const b = bridge()
    try {
      setFound({})
      const ok = b?.startDiscovery?.()
      setScanning(!!ok)
    } catch {
      setScanning(false)
    }
  }

  function stopScan() {
    try {
      bridge()?.stopDiscovery?.()
    } finally {
      setScanning(false)
    }
  }

  function doPair(addr: string) {
    try {
      bridge()?.pair?.(addr)
    } catch {}
  }

  function select(addr: string) {
    try {
      bridge()?.selectPrinter?.(addr)
      setSelected(addr)
    } catch {}
  }

  const pairedList = Object.values(paired)
  const foundList = Object.values(found)

  return (
    <>
      <Button
        variant="outline"
        size="sm"
        className="rounded-[11px] border-0"
        onClick={() => setOpen(true)}
        disabled={!hasBridge}
        aria-label="Drucker auswählen"
      >
        {selectedName ? <Bluetooth className="mr-1 size-4" /> : <WifiOff className="mr-1 size-4" />}
        <span className="hidden md:inline">{selectedName ? selectedName : "Drucker"}</span>
      </Button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle>Drucker verbinden</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" onClick={refreshPaired} disabled={busy}>
                <RefreshCw className="mr-1 size-4" />
                Paired aktualisieren
              </Button>
              {scanning ? (
                <Button variant="outline" size="sm" onClick={stopScan}>
                  <BluetoothSearching className="mr-1 size-4 animate-pulse" />
                  Suche stoppen
                </Button>
              ) : (
                <Button variant="outline" size="sm" onClick={startScan}>
                  <BluetoothSearching className="mr-1 size-4" />
                  Geräte suchen
                </Button>
              )}
            </div>

            <div className="grid gap-2">
              <div className="text-xs font-semibold uppercase text-muted-foreground">Gekoppelt</div>
              {pairedList.length === 0 && <div className="text-sm text-muted-foreground">Keine</div>}
              {pairedList.map((p) => (
                <div key={p.address} className="flex items-center justify-between rounded-md border p-2">
                  <div className="min-w-0">
                    <div className="truncate text-sm font-medium">{p.name || p.address}</div>
                    <div className="truncate text-xs text-muted-foreground">{p.address}</div>
                  </div>
                  <div className="flex items-center gap-2">
                    {selected === p.address ? (
                      <span className="text-emerald-600 inline-flex items-center text-xs">
                        <Check className="mr-1 size-4" /> Ausgewählt
                      </span>
                    ) : (
                      <Button size="sm" onClick={() => select(p.address)} disabled={busy}>
                        Auswählen
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>

            <div className="grid gap-2">
              <div className="text-xs font-semibold uppercase text-muted-foreground">Gefunden</div>
              {foundList.length === 0 && <div className="text-sm text-muted-foreground">Noch keine</div>}
              {foundList.map((p) => (
                <div key={p.address} className="flex items-center justify-between rounded-md border p-2">
                  <div className="min-w-0">
                    <div className="truncate text-sm font-medium">{p.name || p.address}</div>
                    <div className="truncate text-xs text-muted-foreground">{p.address}</div>
                  </div>
                  <div className="flex items-center gap-2">
                    {p.bonded ? (
                      selected === p.address ? (
                        <span className="text-emerald-600 inline-flex items-center text-xs">
                          <Check className="mr-1 size-4" /> Ausgewählt
                        </span>
                      ) : (
                        <Button size="sm" onClick={() => select(p.address)} disabled={busy}>
                          Auswählen
                        </Button>
                      )
                    ) : (
                      <Button size="sm" variant="outline" onClick={() => doPair(p.address)} disabled={busy}>
                        Koppeln
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>
              Schliessen
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

