"use client";

import { useEffect, useRef } from "react";
import QRCode from "qrcode";

type Props = {
  value: string;
  size?: number;
  level?: "L" | "M" | "Q" | "H";
};

export default function QrCodeCanvas({ value, size = 256, level = "M" }: Props) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const opts: QRCode.QRCodeToCanvasOptions = {
      errorCorrectionLevel: level,
      width: size,
      margin: 2,
    } as any;
    QRCode.toCanvas(canvas, value, opts).catch((e) => {
      // eslint-disable-next-line no-console
      console.error("Failed to render QR:", e);
    });
  }, [value, size, level]);

  return <canvas ref={canvasRef} width={size} height={size} />;
}

