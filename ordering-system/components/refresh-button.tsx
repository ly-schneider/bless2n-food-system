import { RefreshCcw } from "lucide-react";
import { Button } from "./ui/button";

export function RefreshButton() {
  const refreshPage = () => {
    window.location.reload();
  };

  return (
    <Button 
      variant="outline" 
      onClick={refreshPage} 
      aria-label="Lock application"
    >
      <RefreshCcw className="h-5 w-5" />
      Neuladen
    </Button>
  );
}