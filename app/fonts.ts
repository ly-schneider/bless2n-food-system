import { Geist } from 'next/font/google';
import localFont from 'next/font/local';

// Define local custom fonts from assets/fonts directory
export const customFont = localFont({
  src: [
    {
      path: '../assets/fonts/FlamaSemicondensed-Medium.otf',
      weight: '400',
      style: 'normal',
    },
  ],
  variable: '--font-custom',
  display: 'swap',
});

// Keep Geist font as a fallback
export const geistSans = Geist({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-geist',
});

// Export all font variables to use in className
export const fontVariables = `${customFont.variable} ${geistSans.variable}`;