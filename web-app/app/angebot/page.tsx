"use client";

import { useState, useEffect, ChangeEvent } from "react";
import { useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { format } from "date-fns";
import { CalendarIcon, CheckCircle, HelpCircle, ArrowLeft, ArrowRight } from "lucide-react";
import { de } from "date-fns/locale";
import {
  Card,
  CardContent, CardHeader,
  CardTitle
} from "@/components/ui/card";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import Link from "next/link";

// Define TypeScript types for form data
interface FormData {
  plan: string;
  checkoutStations: string;
  location: string;
  startDate: Date | null;
  endDate: Date | null;
  checkoutType: string;
  name: string;
  email: string;
  phone: string;
}

// Plan type
type PlanType = "lite" | "pro";

// CheckoutType
type CheckoutType = "dual-display" | "self-checkout";

export default function AngebotPage() {
  const searchParams = useSearchParams();
  
  // Form state
  const [step, setStep] = useState<number>(1);
  const [formData, setFormData] = useState<FormData>({
    plan: searchParams.get("plan") || "",
    checkoutStations: "",
    location: "",
    startDate: null,
    endDate: null,
    checkoutType: "",
    name: "",
    email: "",
    phone: "",
  });
  
  // Success state
  const [isSuccess, setIsSuccess] = useState<boolean>(false);
  const [offerId, setOfferId] = useState<string>("");

  useEffect(() => {
    // Check for plan parameter
    if (formData.plan) {
      setStep(2);
    } 
    // Check for angebot parameter
    else {
      const angebotParam = searchParams.get("angebot");
      if (angebotParam === "lite" || angebotParam === "pro") {
        // Update the plan and this will trigger the effect again to move to step 2
        setFormData(prev => ({ ...prev, plan: angebotParam }));
      }
    }
  }, [formData.plan, searchParams]);

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleRadioChange = (name: keyof FormData, value: string) => {
    setFormData(prev => ({ ...prev, [name]: value }));
    
    // If we're on step 1 and changing the plan, clean up URL params
    // to avoid confusion with the original selection from URL
    if (name === "plan" && step === 1 && searchParams.has("angebot")) {
      // Create a new URL with current path but without the angebot param
      const url = new URL(window.location.href);
      url.searchParams.delete("angebot");
      
      // Replace current history state with the cleaned URL
      window.history.replaceState({}, "", url.toString());
    }
  };

  const handleDateChange = (name: 'startDate' | 'endDate', date: Date | null) => {
    setFormData(prev => ({ ...prev, [name]: date }));
  };

  const nextStep = () => {
    setStep(prev => prev + 1);
    window.scrollTo(0, 0);
  };

  const prevStep = () => {
    // Just handle step transition without URL manipulation
    setStep(prev => prev - 1);
    window.scrollTo(0, 0);
  };

  const handleSubmit = () => {
    // In a real app, you would send this data to your server
    console.log("Submitted form data:", formData);
    
    // Generate a random offer ID
    const randomId = Math.random().toString(36).substring(2, 10).toUpperCase();
    setOfferId(`RENTRO-${randomId}`);
    setIsSuccess(true);
  };

  // Validation for the current step
  const isStepValid = (): boolean => {
    switch (step) {
      case 1:
        return !!formData.plan;
      case 2:
        return !!formData.location && !!formData.startDate && !!formData.endDate && !!formData.checkoutType;
      case 3:
        return !!formData.name && !!formData.email;
      default:
        return true;
    }
  };

  // Render step content
  const renderStep = () => {
    switch (step) {
      case 1:
        return renderPlanSelection();
      case 2:
        return renderEventDetails();
      case 3:
        return renderContactInfo();
      case 4:
        return renderSummary();
      default:
        return null;
    }
  };

  // Step 1: Plan selection
  const renderPlanSelection = () => (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Wählen Sie Ihren Plan</h2>
      <RadioGroup 
        value={formData.plan} 
        onValueChange={(value: string) => handleRadioChange("plan", value)}
        className="grid grid-cols-1 md:grid-cols-2 gap-4"
      >
        <div className={`border-2 p-6 rounded-lg ${formData.plan === "lite" ? "border-foreground" : "border-gray-200"}`}>
          <RadioGroupItem value="lite" id="lite" className="sr-only" />
          <Label htmlFor="lite" className="flex flex-col cursor-pointer">
            <span className="text-xl font-bold mb-2">Event Lite</span>
            <span className="text-2xl font-bold mb-2">CHF 390</span>
            <span className="text-sm text-gray-500 mb-4">für 3 Tage • CHF 120 je Zusatztag</span>
            <span className="text-gray-700">Für kleine & mittlere Anlässe – inkl. persönlichem Vor-Ort Setup</span>
          </Label>
        </div>
        
        <div className={`border-2 p-6 rounded-lg ${formData.plan === "pro" ? "border-foreground" : "border-gray-200"}`}>
          <RadioGroupItem value="pro" id="pro" className="sr-only" />
          <Label htmlFor="pro" className="flex flex-col cursor-pointer">
            <span className="text-xl font-bold mb-2">Event Pro</span>
            <span className="text-2xl font-bold mb-2">CHF 780</span>
            <span className="text-sm text-gray-500 mb-4">für 3 Tage • CHF 200 je Zusatztag</span>
            <span className="text-gray-700">Für Festivals & Grossanlässe – inkl. Premium-Features</span>
          </Label>
        </div>
      </RadioGroup>
    </div>
  );

  // Step 2: Event details
  const renderEventDetails = () => (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Event Details</h2>
      
      <div>
        <Label htmlFor="checkoutStations" className="block mb-2">
          Anzahl Kassen-Stationen (optional)
        </Label>
        <Input 
          id="checkoutStations"
          name="checkoutStations"
          type="number"
          min="1"
          placeholder="z.B. 3"
          value={formData.checkoutStations}
          onChange={handleInputChange}
          className="w-full"
        />
        <p className="text-sm text-gray-500 mt-1">
          Lassen Sie dieses Feld leer, wenn Sie unsicher sind. Wir beraten Sie gerne.
        </p>
      </div>

      <div>
        <Label htmlFor="location" className="block mb-2">
          Veranstaltungsort *
        </Label>
        <Input 
          id="location"
          name="location"
          required
          placeholder="z.B. Zürich, Eventlocation XYZ"
          value={formData.location}
          onChange={handleInputChange}
          className="w-full"
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <Label className="block mb-2">Event Startdatum *</Label>
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                className={`w-full justify-start text-left font-normal ${!formData.startDate && "text-muted-foreground"}`}
              >
                <CalendarIcon className="mr-2 h-4 w-4" />
                {formData.startDate ? format(formData.startDate, "PPP", { locale: de }) : "Datum auswählen"}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
              <Calendar
                mode="single"
                selected={formData.startDate ?? undefined}
                onSelect={(date) => handleDateChange("startDate", date ?? null)}
                initialFocus
                locale={de}
              />
            </PopoverContent>
          </Popover>
        </div>
        
        <div>
          <Label className="block mb-2">Event Enddatum *</Label>
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                className={`w-full justify-start text-left font-normal ${!formData.endDate && "text-muted-foreground"}`}
              >
                <CalendarIcon className="mr-2 h-4 w-4" />
                {formData.endDate ? format(formData.endDate, "PPP", { locale: de }) : "Datum auswählen"}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
              <Calendar
                mode="single"
                selected={formData.endDate ?? undefined}
                onSelect={(date) => handleDateChange("endDate", date ?? null)}
                disabled={(date) => (formData.startDate ? date < formData.startDate : false)}
                initialFocus
                locale={de}
              />
            </PopoverContent>
          </Popover>
        </div>
      </div>

      <div>
        <div className="flex items-center mb-2">
          <Label className="block">Kassen-Typ *</Label>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <HelpCircle className="h-4 w-4 ml-2 text-gray-500" />
              </TooltipTrigger>
              <TooltipContent>
                <p className="max-w-xs">
                  <strong>Dual-Display:</strong> Ein Bildschirm für den Kassierer, ein Bildschirm für den Kunden.<br />
                  <strong>Self-Checkout:</strong> Kunden können ihre Bestellung selbst eingeben und bezahlen.
                </p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
        
        <RadioGroup 
          value={formData.checkoutType} 
          onValueChange={(value: string) => handleRadioChange("checkoutType", value)}
          className="grid grid-cols-1 md:grid-cols-2 gap-4"
        >
          <div className={`border-2 p-4 rounded-lg ${formData.checkoutType === "dual-display" ? "border-foreground" : "border-gray-200"}`}>
            <RadioGroupItem value="dual-display" id="dual-display" className="sr-only" />
            <Label htmlFor="dual-display" className="flex flex-col cursor-pointer">
              <span className="font-bold">Dual-Display</span>
              <span className="text-sm text-gray-500">Separate Ansichten für Kassierer und Kunde</span>
            </Label>
          </div>
          
          <div className={`border-2 p-4 rounded-lg ${formData.checkoutType === "self-checkout" ? "border-foreground" : "border-gray-200"}`}>
            <RadioGroupItem value="self-checkout" id="self-checkout" className="sr-only" />
            <Label htmlFor="self-checkout" className="flex flex-col cursor-pointer">
              <span className="font-bold">Self-Checkout</span>
              <span className="text-sm text-gray-500">Kunden wählen und zahlen selbstständig</span>
            </Label>
          </div>
        </RadioGroup>
      </div>
    </div>
  );

  // Step 3: Contact Information
  const renderContactInfo = () => (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Kontaktinformationen</h2>
      
      <div>
        <Label htmlFor="name" className="block mb-2">
          Name *
        </Label>
        <Input 
          id="name"
          name="name"
          required
          placeholder="Ihr Name"
          value={formData.name}
          onChange={handleInputChange}
          className="w-full"
        />
      </div>
      
      <div>
        <Label htmlFor="email" className="block mb-2">
          E-Mail Adresse *
        </Label>
        <Input 
          id="email"
          name="email"
          type="email"
          required
          placeholder="ihre.email@beispiel.ch"
          value={formData.email}
          onChange={handleInputChange}
          className="w-full"
        />
      </div>
      
      <div>
        <Label htmlFor="phone" className="block mb-2">
          Telefonnummer (optional)
        </Label>
        <Input 
          id="phone"
          name="phone"
          placeholder="+41 XX XXX XX XX"
          value={formData.phone}
          onChange={handleInputChange}
          className="w-full"
        />
      </div>
    </div>
  );

  // Step 4: Summary
  const renderSummary = () => {
    // Calculate total price
    const basePlanPrice = formData.plan === "lite" ? 390 : 780;
    const baseDays = 3;
    const additionalDayPrice = formData.plan === "lite" ? 120 : 200;
    
    // Calculate days between start and end date
    const startDate = formData.startDate as Date; // We know it's not null because of validation
    const endDate = formData.endDate as Date; // We know it's not null because of validation
    const diffTime = Math.abs(endDate.getTime() - startDate.getTime());
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1; // +1 because we count both start and end day
    
    const additionalDays = Math.max(0, diffDays - baseDays);
    const additionalDaysPrice = additionalDays * additionalDayPrice;
    
    const totalPrice = basePlanPrice + additionalDaysPrice;

    return (
      <div className="space-y-6">
        <h2 className="text-2xl font-bold">Zusammenfassung</h2>
        
        <Card>
          <CardHeader>
            <CardTitle>Ihr gewählter Plan</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <p className="font-bold">
                  {formData.plan === "lite" ? "Event Lite" : "Event Pro"}
                </p>
                <p className="text-gray-500">
                  {formData.plan === "lite" 
                    ? "CHF 390 für 3 Tage (CHF 120 je Zusatztag)" 
                    : "CHF 780 für 3 Tage (CHF 200 je Zusatztag)"}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader>
            <CardTitle>Event Details</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {formData.checkoutStations && (
                <div className="flex justify-between">
                  <span className="text-gray-500">Anzahl Kassen-Stationen:</span>
                  <span>{formData.checkoutStations}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-gray-500">Veranstaltungsort:</span>
                <span>{formData.location}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Zeitraum:</span>
                <span>
                  {formData.startDate && format(formData.startDate, "dd.MM.yyyy", { locale: de })}{" "}-{" "} 
                  {formData.endDate && format(formData.endDate, "dd.MM.yyyy", { locale: de })}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Dauer:</span>
                <span>{diffDays} Tage</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Kassen-Typ:</span>
                <span>{formData.checkoutType === "dual-display" ? "Dual-Display" : "Self-Checkout"}</span>
              </div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader>
            <CardTitle>Kontaktinformationen</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-gray-500">Name:</span>
                <span>{formData.name}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">E-Mail:</span>
                <span>{formData.email}</span>
              </div>
              {formData.phone && (
                <div className="flex justify-between">
                  <span className="text-gray-500">Telefon:</span>
                  <span>{formData.phone}</span>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
        
        <Card className="bg-foreground text-white">
          <CardHeader>
            <CardTitle>Kostenübersicht</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span>Basispreis ({baseDays} Tage):</span>
                <span>CHF {basePlanPrice.toFixed(2)}</span>
              </div>
              {additionalDays > 0 && (
                <div className="flex justify-between">
                  <span>Zusatztage ({additionalDays} × CHF {additionalDayPrice}):</span>
                  <span>CHF {additionalDaysPrice.toFixed(2)}</span>
                </div>
              )}
              <div className="h-px bg-white/20 my-2"></div>
              <div className="flex justify-between text-lg font-bold">
                <span>Gesamtbetrag:</span>
                <span>CHF {totalPrice.toFixed(2)}</span>
              </div>
              <p className="text-sm text-white/70 mt-2">
                Alle Preise verstehen sich exkl. MwSt.
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  };

  // Success screen
  const renderSuccess = () => (
    <div className="text-center space-y-6 max-w-2xl mx-auto">
      <div className="mx-auto w-16 h-16 bg-green-100 rounded-full flex items-center justify-center">
        <CheckCircle className="w-8 h-8 text-green-600" />
      </div>
      
      <h1 className="text-3xl font-bold">Vielen Dank für Ihre Anfrage!</h1>
      
      <p className="text-lg text-gray-600">
        Wir haben Ihre Angebotsanfrage erhalten und werden uns in Kürze mit Ihnen in Verbindung setzen.
      </p>
      
      <div className="bg-muted p-4 rounded-lg">
        <p className="text-sm text-gray-500">Ihre Angebots-ID</p>
        <p className="text-xl font-mono">{offerId}</p>
      </div>
      
      <p className="text-gray-600">
        Eine Bestätigung wurde an {formData.email} gesendet.
      </p>
      
      <div className="pt-6">
        <Link href="/">
          <Button>
            Zurück zur Startseite
          </Button>
        </Link>
      </div>
    </div>
  );

  // Progress indicator
  const renderProgress = () => (
    <div className="mb-8">
      <div className="flex justify-between">
        {[1, 2, 3, 4].map((stepNumber) => (
          <div 
            key={stepNumber} 
            className={`flex flex-col items-center ${stepNumber < step ? "text-foreground" : stepNumber === step ? "text-foreground" : "text-gray-300"}`}
          >
            <div 
              className={`w-8 h-8 rounded-full flex items-center justify-center mb-1 ${
                stepNumber < step 
                  ? "bg-foreground text-white" 
                  : stepNumber === step 
                    ? "border-2 border-foreground" 
                    : "border-2 border-gray-300"
              }`}
            >
              {stepNumber < step ? "✓" : stepNumber}
            </div>
            <span className="text-xs hidden md:block">
              {stepNumber === 1 ? "Plan" : 
               stepNumber === 2 ? "Details" : 
               stepNumber === 3 ? "Kontakt" : "Übersicht"}
            </span>
          </div>
        ))}
      </div>
      <div className="mt-2 grid grid-cols-3 gap-0">
        {[1, 2, 3].map((lineNumber) => (
          <div 
            key={lineNumber}
            className={`h-1 ${lineNumber < step ? "bg-foreground" : "bg-gray-300"}`}
          />
        ))}
      </div>
    </div>
  );

  return (
    <div className="pt-32 pb-20">
      <div className="container mx-auto px-4 md:px-6">
        {!isSuccess ? (
          <>
            {renderProgress()}
            
            <div className="max-w-3xl mx-auto">
              {renderStep()}
              
              <div className="mt-12 flex justify-between">
                {step > 1 && (
                  <Button
                    variant="outline"
                    onClick={prevStep}
                    className="flex items-center"
                  >
                    <ArrowLeft className="mr-2 h-4 w-4" /> Zurück
                  </Button>
                )}
                
                {step < 4 ? (
                  <Button
                    onClick={nextStep}
                    disabled={!isStepValid()}
                    className="ml-auto flex items-center"
                    variant={isStepValid() ? "default" : "outline"}
                  >
                    Weiter <ArrowRight className="ml-2 h-4 w-4" />
                  </Button>
                ) : (
                  <Button
                    onClick={handleSubmit}
                    className="ml-auto"
                    variant={isStepValid() ? "default" : "outline"}
                  >
                    Anfrage absenden
                  </Button>
                )}
              </div>
            </div>
          </>
        ) : (
          renderSuccess()
        )}
      </div>
    </div>
  );
}
