'use client';

import { useEffect, useState, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { createClientComponentClient } from '@supabase/auth-helpers-nextjs';
import { FlipHorizontal } from 'lucide-react';
import { Html5Qrcode, Html5QrcodeScanner, Html5QrcodeSupportedFormats } from 'html5-qrcode';

interface OrderItem {
  product_name: string;
  quantity: number;
  price: number;
  color_hex: string;
  emoji: string;
}

interface Order {
  id: string;
  items: OrderItem[];
  total: number;
  created_at: string;
  order_number?: string;
}

export default function StationPage() {
  const [order, setOrder] = useState<Order | null>(null);
  const [error, setError] = useState<string>('');
  const [lastScannedCode, setLastScannedCode] = useState<string>('');
  const [audio, setAudio] = useState<HTMLAudioElement | null>(null);
  const [isCameraActive, setIsCameraActive] = useState(false);
  const [currentCameraIndex, setCurrentCameraIndex] = useState(0);
  const [currentCameraLabel, setCurrentCameraLabel] = useState<string>('');
  const [cameras, setCameras] = useState<MediaDeviceInfo[]>([]);
  const [scanner, setScanner] = useState<any>(null);
  const [html5QrCode, setHtml5QrCode] = useState<any>(null);
  const [isScanning, setIsScanning] = useState(true);
  const [orderJustScanned, setOrderJustScanned] = useState(false);
  const supabase = createClientComponentClient();

  // Initialize audio
  useEffect(() => {
    const initAudio = async () => {
      try {
        const sound = new Audio('/sounds/beep.mp3');
        // Preload the audio
        await sound.load();
        setAudio(sound);
        console.log('Audio initialized successfully');
      } catch (err) {
        console.error('Failed to initialize audio:', err);
      }
    };

    initAudio();
  }, []);

  const playSuccessSound = useCallback(async () => {
    try {
      if (audio) {
        audio.currentTime = 0;
        await audio.play();
        console.log('Sound played successfully');
      } else {
        console.warn('Audio not initialized');
      }
    } catch (err) {
      console.error('Error playing sound:', err);
    }
  }, [audio]);

  useEffect(() => {
    const setupScanner = async () => {
      try {
        // Dynamic import of QR Scanner to avoid SSR issues
        const { Html5QrcodeScanner, Html5Qrcode } = await import('html5-qrcode');
        
        // Get available cameras
        const devices = await Html5Qrcode.getCameras();
        // Convert CameraDevice[] to MediaDeviceInfo[]
        const mediaDevices = devices.map(device => ({
          deviceId: device.id,
          groupId: '',
          kind: 'videoinput',
          label: device.label,
          toJSON: () => ({ deviceId: device.id, groupId: '', kind: 'videoinput', label: device.label })
        } as MediaDeviceInfo));
        setCameras(mediaDevices);
        
        if (devices && devices.length > 0) {
          setCurrentCameraLabel(devices[0].label);
          const html5QrCode = new Html5Qrcode("reader");
          setHtml5QrCode(html5QrCode);

          const verbose = false;
          const newScanner = new Html5QrcodeScanner('reader', {
            qrbox: { width: 250, height: 250 },
            fps: 10,
            rememberLastUsedCamera: false,
            aspectRatio: 1.33,
            showTorchButtonIfSupported: true,
            formatsToSupport: [ Html5QrcodeSupportedFormats.QR_CODE ],
            videoConstraints: {
              width: { min: 320, ideal: 1280, max: 1920 },
              height: { min: 240, ideal: 720, max: 1080 },
              facingMode: "environment"
            }
          }, verbose);

          setScanner(newScanner);
          newScanner.render(onScanSuccess, onScanError);

          // Hide unnecessary UI elements but keep camera select
          const elements = [
            'reader__dashboard_section_swaplink'
          ];
          elements.forEach(id => {
            const element = document.getElementById(id);
            if (element) {
              element.style.display = 'none';
            }
          });

          // Style the camera section
          const dashboardSection = document.getElementById('reader__dashboard_section_csr');
          if (dashboardSection) {
            dashboardSection.style.display = 'none';
          }

          return () => {
            html5QrCode.stop().catch(console.error);
            newScanner.clear().catch(console.error);
          };
        }
      } catch (err) {
        console.error('Failed to initialize scanner:', err);
        setError('Failed to initialize scanner');
      }
    };

    setupScanner();
  }, []);

  useEffect(() => {
    const handleScanningState = async () => {
      if (!html5QrCode) return;

      try {
        if (!isScanning) {
          await html5QrCode.stop();
          setIsCameraActive(false);
        } else {
          if (cameras && cameras.length > 0) {
            const cameraId = cameras[currentCameraIndex].deviceId;
            await html5QrCode.start(
              cameraId,
              {
                fps: 10,
                qrbox: { width: 250, height: 250 },
                aspectRatio: 1.33
              },
              onScanSuccess,
              onScanError
            );
            setIsCameraActive(true);
          }
        }
      } catch (err) {
        console.error('Error handling scanning state:', err);
      }
    };

    handleScanningState();
  }, [isScanning, html5QrCode, currentCameraIndex]);

  const startCamera = () => {
    if (!scanner) return;
    
    const startButton = document.getElementById('reader__dashboard_section_csr');
    if (startButton) {
      const startScanButton = startButton.querySelector('button');
      if (startScanButton) {
        startScanButton.click();
        setIsCameraActive(true);

        // Get current camera label
        navigator.mediaDevices.enumerateDevices().then(devices => {
          const cameras = devices.filter(device => device.kind === 'videoinput');
          if (cameras[currentCameraIndex]) {
            setCurrentCameraLabel(cameras[currentCameraIndex].label || `Camera ${currentCameraIndex + 1}`);
          }
        });
      }
    }
  };

  const switchCamera = async () => {
    if (!scanner) return;
    
    try {
      // First pause the current scanner
      await html5QrCode.stop();
      
      if (cameras.length > 1) {
        const nextIndex = (currentCameraIndex + 1) % cameras.length;
        setCurrentCameraIndex(nextIndex);
        
        // Update camera label
        setCurrentCameraLabel(cameras[nextIndex].label || `Camera ${nextIndex + 1}`);

        // Get the camera selection element and trigger change
        const cameraSelect = document.getElementById('reader__camera_selection');
        if (cameraSelect instanceof HTMLSelectElement) {
          cameraSelect.selectedIndex = nextIndex;
          // Create and dispatch change event
          const event = new Event('change', { bubbles: true });
          cameraSelect.dispatchEvent(event);
        }

        // Resume scanning
        await html5QrCode.start(
          cameras[nextIndex].deviceId,
          {
            fps: 10,
            qrbox: { width: 250, height: 250 },
            aspectRatio: 1.33
          },
          onScanSuccess,
          onScanError
        );
      }
    } catch (err) {
      console.error('Failed to switch camera:', err);
      // Try to restart the current camera if switch fails
      const startButton = document.getElementById('reader__dashboard_section_csr');
      if (startButton) {
        const startScanButton = startButton.querySelector('button');
        if (startScanButton) {
          startScanButton.click();
        }
      }
    }
  };

  const onScanSuccess = async (decodedText: string) => {
    if (!isScanning) return;
    
    setLastScannedCode(decodedText);
    
    try {
      // Stop scanning immediately
      if (html5QrCode) {
        await html5QrCode.stop();
      }

      const { data, error: functionError } = await supabase.functions.invoke('get_order', {
        body: { order_id: decodedText }
      });

      if (functionError) throw functionError;

      if (!data) {
        setError('Order not found');
        setOrder(null);
        // Restart scanning if order not found
        if (html5QrCode) {
          await html5QrCode.start(
            cameras[currentCameraIndex].deviceId,
            {
              fps: 10,
              qrbox: { width: 250, height: 250 },
              aspectRatio: 1.33
            },
            onScanSuccess,
            onScanError
          );
        }
        return;
      }

      // Play sound first to ensure it plays
      await playSuccessSound();
      console.log('Scan successful, sound should have played');
      
      setOrder(data as Order);
      setError('');
      setIsScanning(false);
      setOrderJustScanned(true);

      // Start the background transition
      setTimeout(async () => {
        setOrderJustScanned(false);
        if (html5QrCode) {
          await html5QrCode.start(
            cameras[currentCameraIndex].deviceId,
            {
              fps: 10,
              qrbox: { width: 250, height: 250 },
              aspectRatio: 1.33
            },
            onScanSuccess,
            onScanError
          );
          setIsScanning(true);
        }
      }, 7000);

    } catch (err) {
      console.error('Failed to fetch order:', err);
      setError('Failed to fetch order details');
      setOrder(null);
    }
  };

  const onScanError = (err: any) => {
    console.warn(err);
  };

  return (
    <div className={`min-h-screen p-8 transition-all duration-[7000ms] ${
      orderJustScanned ? 'bg-green-500' : 'bg-white'
    }`}>
      {/* Scanner Overlay */}
      <div className="absolute top-4 left-4 w-full max-w-[320px] aspect-video bg-black rounded-lg overflow-hidden shadow-xl">
        <div id="reader" className="w-full h-full"></div>
        {!isCameraActive && (
          <button
            onClick={startCamera}
            className="absolute inset-0 w-full h-full flex items-center justify-center bg-black/50 text-white hover:bg-black/60 transition-colors"
          >
            Kamera starten
          </button>
        )}
        {isCameraActive && (
          <>
            <button
              onClick={switchCamera}
              className="absolute bottom-4 right-4 p-2 bg-black/50 text-white rounded-full hover:bg-black/60 transition-colors"
              title="Kamera wechseln"
            >
              <FlipHorizontal className="w-6 h-6" />
            </button>
            <div className="absolute bottom-4 left-4 right-16 px-2 py-1 bg-black/50 text-white text-sm truncate rounded">
              {currentCameraLabel}
            </div>
          </>
        )}
      </div>

      {/* Order Details */}
      <div className="pt-[280px] md:pt-8 md:ml-[340px]">
        <Card 
          style={{
            backgroundColor: orderJustScanned ? 'rgb(187 247 208)' : undefined,
            transition: 'background-color 7s ease-out'
          }}
        >
          <CardHeader className="bg-inherit">
            <CardTitle className="text-2xl">Bestelldetails</CardTitle>
          </CardHeader>
          <CardContent className="bg-inherit">
            {order ? (
              <ScrollArea className="h-[calc(100vh-400px)] w-full pr-4">
                <div className="space-y-6">
                  <div>
                    <h3 className="text-2xl font-bold mb-2">#{order.order_number || order.id}</h3>
                    <p className="text-lg text-gray-500">
                      Erstellt: {new Date(order.created_at).toLocaleString('de-DE')}
                    </p>
                  </div>

                  <div className="space-y-4">
                    {order.items.map((item, index) => (
                      <div 
                        key={index} 
                        style={{ 
                          backgroundColor: isScanning 
                            ? '#f3f4f6' // gray-100
                            : `${item.color_hex}20`
                        }}
                        className="p-6 rounded-lg"
                      >
                        <div className="flex items-center gap-6">
                          <div className="text-4xl">{item.emoji}</div>
                          <div className="flex-1">
                            <p className="text-3xl font-bold mb-2">{item.product_name}</p>
                            <div className="flex items-center">
                              <span className="text-4xl font-black bg-white px-4 py-2 rounded-lg border-2 border-black">
                                {item.quantity}x
                              </span>
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </ScrollArea>
            ) : (
              <div className="text-center py-12 text-xl text-gray-500">
                Bitte scannen Sie einen QR-Code, um die Bestelldetails anzuzeigen
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
