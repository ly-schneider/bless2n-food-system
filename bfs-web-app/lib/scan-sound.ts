type SoundType = "success" | "warning" | "error"

let audioCtx: AudioContext | null = null

function getCtx(): AudioContext | null {
  if (typeof window === "undefined") return null
  try {
    if (!audioCtx) {
      const Ctor =
        window.AudioContext || (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
      if (!Ctor) return null
      audioCtx = new Ctor()
    }
    if (audioCtx.state === "suspended") {
      audioCtx.resume().catch(() => {})
    }
    return audioCtx
  } catch {
    return null
  }
}

function beep(
  ctx: AudioContext,
  freq: number,
  startOffset: number,
  duration: number,
  type: OscillatorType,
  peak = 0.18
) {
  const osc = ctx.createOscillator()
  const gain = ctx.createGain()
  osc.type = type
  osc.frequency.setValueAtTime(freq, ctx.currentTime + startOffset)
  gain.gain.setValueAtTime(0.0001, ctx.currentTime + startOffset)
  gain.gain.exponentialRampToValueAtTime(peak, ctx.currentTime + startOffset + 0.01)
  gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + startOffset + duration)
  osc.connect(gain)
  gain.connect(ctx.destination)
  osc.start(ctx.currentTime + startOffset)
  osc.stop(ctx.currentTime + startOffset + duration + 0.02)
}

export function playScanSound(type: SoundType): void {
  const ctx = getCtx()
  if (!ctx) return
  try {
    if (type === "success") {
      beep(ctx, 880, 0, 0.09, "sine", 0.22)
      beep(ctx, 1318, 0.09, 0.16, "sine", 0.22)
    } else if (type === "warning") {
      beep(ctx, 520, 0, 0.14, "sine", 0.2)
      beep(ctx, 520, 0.18, 0.14, "sine", 0.2)
    } else {
      beep(ctx, 180, 0, 0.18, "square", 0.16)
      beep(ctx, 140, 0.2, 0.22, "square", 0.16)
    }
  } catch {
    // ignore audio failures
  }
}

export function primeScanAudio(): void {
  getCtx()
}
