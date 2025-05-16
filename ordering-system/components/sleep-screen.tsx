"use client";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Moon } from "lucide-react";

interface SleepScreenProps {
  isAsleep: boolean;
  onWakeUp: () => void;
}

export function SleepScreen({ isAsleep, onWakeUp }: SleepScreenProps) {
  return (
    <Dialog open={isAsleep} onOpenChange={() => {}}>
      <DialogHeader className="text-center">
        <DialogTitle className="text-2xl flex items-center justify-center gap-2">
          <Moon className="h-6 w-6" />
          Bildschirm im Ruhemodus
        </DialogTitle>
      </DialogHeader>
      <DialogContent
        className="rounded-none flex flex-col items-center justify-center bg-black text-white max-w-none w-full min-h-screen [&>p]:opacity-50"
        onClick={onWakeUp}
      >
        <p className="text-2xl font-medium mb-4">Bildschirm im Ruhemodus</p>
        <p className="text-lg">Berühren Sie den Bildschirm, um fortzufahren</p>
      </DialogContent>
    </Dialog>
  );
}
