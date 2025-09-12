"use client";

import { useEffect, useMemo, useState } from "react";
import { createClient } from "@/utils/supabase/client";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import QrCodeCanvas from "@/components/qr-code-canvas";

type Order = {
  id: string;
  created_at: string;
  total?: number;
};

export default function QrCodesPage() {
  const supabase = createClient();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>("");
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const {
          data: { user },
        } = await supabase.auth.getUser();
        if (!user) {
          setOrders([]);
          setError("Nicht angemeldet");
          return;
        }

        const { data, error } = await supabase
          .from("orders")
          .select("id, created_at, total")
          .eq("admin_id", user.id)
          .order("created_at", { ascending: false })
          .limit(100);
        if (error) throw error;
        setOrders((data as Order[]) || []);
        setError("");
      } catch (e: any) {
        // eslint-disable-next-line no-console
        console.error(e);
        setError("Bestellungen konnten nicht geladen werden");
      } finally {
        setLoading(false);
      }
    };
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const list = useMemo(() => orders, [orders]);

  return (
    <div className="px-5 pb-5">
      <div className="flex items-center justify-between py-4">
        <h1 className="text-xl font-semibold">QR-Codes</h1>
      </div>
      {loading ? (
        <div className="text-sm text-muted-foreground">Lade Bestellungen…</div>
      ) : error ? (
        <div className="text-sm text-destructive">{error}</div>
      ) : list.length === 0 ? (
        <div className="text-sm text-muted-foreground">Keine Bestellungen gefunden</div>
      ) : (
        <ul className="space-y-2">
          {list.map((o) => (
            <li
              key={o.id}
              className="flex items-center justify-between rounded-md border p-3 bg-background"
            >
              <div className="flex flex-col">
                <span className="font-mono text-sm">{o.id}</span>
                <span className="text-xs text-muted-foreground">
                  {new Date(o.created_at).toLocaleString("de-CH")}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <Button size="sm" onClick={() => setSelectedOrder(o)}>
                  QR anzeigen
                </Button>
              </div>
            </li>
          ))}
        </ul>
      )}

      <Dialog open={!!selectedOrder} onOpenChange={() => setSelectedOrder(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>QR-Code für Bestellung</DialogTitle>
          </DialogHeader>
          <div className="flex items-center justify-center py-2">
            {selectedOrder && (
              <QrCodeCanvas value={selectedOrder.id} size={256} level="M" />
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setSelectedOrder(null)}>
              Schliessen
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

