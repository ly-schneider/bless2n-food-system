"use client"

import React, { createContext, ReactNode, useContext, useEffect, useReducer } from 'react';
import { ProductDTO } from '@/types';
import { Cart, CartContextType, CartItem, CartItemConfiguration } from '@/types/cart';

const CART_STORAGE_KEY = 'bfs-cart';

function saveCartToStorage(cart: Cart) {
  try {
    localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(cart));
  } catch (error) {
    console.warn('Failed to save cart to localStorage:', error);
  }
}

function loadCartFromStorage(): Cart | null {
  try {
    const storedCart = localStorage.getItem(CART_STORAGE_KEY);
    if (storedCart) {
      return JSON.parse(storedCart) as Cart;
    }
  } catch (error) {
    console.warn('Failed to load cart from localStorage:', error);
  }
  return null;
}

type CartAction =
  | { type: 'ADD_TO_CART'; product: ProductDTO; configuration?: CartItemConfiguration }
  | { type: 'REMOVE_FROM_CART'; itemId: string }
  | { type: 'UPDATE_QUANTITY'; itemId: string; quantity: number }
  | { type: 'REPLACE_ITEM'; oldItemId: string; product: ProductDTO; configuration?: CartItemConfiguration }
  | { type: 'CLEAR_CART' }
  | { type: 'LOAD_FROM_STORAGE'; cart: Cart };

function generateCartItemId(product: ProductDTO, configuration?: CartItemConfiguration): string {
  const configStr = configuration ? JSON.stringify(configuration) : '';
  return `${product.id}-${configStr}`;
}

function calculateItemPrice(product: ProductDTO): number {
  // All products (both simple and menu) use their defined priceCents
  return product.priceCents;
}

function cartReducer(state: Cart, action: CartAction): Cart {
  let newState: Cart;
  
  switch (action.type) {
    case 'ADD_TO_CART': {
      const itemId = generateCartItemId(action.product, action.configuration);
      const existingItemIndex = state.items.findIndex(item => item.id === itemId);
      
      if (existingItemIndex >= 0) {
        const updatedItems = [...state.items];
        const existingItem = updatedItems[existingItemIndex];
        if (existingItem) {
          updatedItems[existingItemIndex] = {
            id: existingItem.id,
            product: existingItem.product,
            configuration: existingItem.configuration,
            totalPriceCents: existingItem.totalPriceCents,
            quantity: existingItem.quantity + 1,
          };
        }
        
        const totalCents = updatedItems.reduce((sum, item) => sum + (item.totalPriceCents * item.quantity), 0);
        
        newState = {
          ...state,
          items: updatedItems,
          totalCents,
        };

        break;
      }
      
      const newItem: CartItem = {
        id: itemId,
        product: action.product,
        quantity: 1,
        configuration: action.configuration,
        totalPriceCents: calculateItemPrice(action.product),
      };
      
      const updatedItems = [...state.items, newItem];
      const totalCents = updatedItems.reduce((sum, item) => sum + (item.totalPriceCents * item.quantity), 0);
      
      newState = {
        ...state,
        items: updatedItems,
        totalCents,
      };
      
      break;
    }
    
    case 'REMOVE_FROM_CART': {
      const updatedItems = state.items.filter(item => item.id !== action.itemId);
      const totalCents = updatedItems.reduce((sum, item) => sum + (item.totalPriceCents * item.quantity), 0);
      
      newState = {
        ...state,
        items: updatedItems,
        totalCents,
      };
      
      break;
    }
    
    case 'UPDATE_QUANTITY': {
      if (action.quantity <= 0) {
        return cartReducer(state, { type: 'REMOVE_FROM_CART', itemId: action.itemId });
      }
      
      const updatedItems = state.items.map(item =>
        item.id === action.itemId ? { ...item, quantity: action.quantity } : item
      );
      
      const totalCents = updatedItems.reduce((sum, item) => sum + (item.totalPriceCents * item.quantity), 0);
      
      newState = {
        ...state,
        items: updatedItems,
        totalCents,
      };
      
      break;
    }

    case 'REPLACE_ITEM': {
      const oldIndex = state.items.findIndex((i) => i.id === action.oldItemId);
      if (oldIndex === -1) {
        // If old item not found, fallback to add behavior
        return cartReducer(state, { type: 'ADD_TO_CART', product: action.product, configuration: action.configuration });
      }

      const oldItem = state.items[oldIndex];
      const newItemId = generateCartItemId(action.product, action.configuration);

      // Remove old item first
      const itemsWithoutOld = state.items.filter((_, idx) => idx !== oldIndex);

      // Check if new item already exists to merge quantities
      const existingNewIndex = itemsWithoutOld.findIndex((i) => i.id === newItemId);
      let updatedItems: CartItem[];

      if (existingNewIndex >= 0) {
        // Merge quantities into existing new item
        updatedItems = itemsWithoutOld.map((it, idx) =>
          idx === existingNewIndex ? { ...it, quantity: it.quantity + oldItem.quantity } : it
        );
      } else {
        const newItem: CartItem = {
          id: newItemId,
          product: action.product,
          configuration: action.configuration,
          quantity: oldItem.quantity,
          totalPriceCents: calculateItemPrice(action.product),
        };
        updatedItems = [...itemsWithoutOld, newItem];
      }

      const totalCents = updatedItems.reduce((sum, item) => sum + item.totalPriceCents * item.quantity, 0);

      newState = {
        ...state,
        items: updatedItems,
        totalCents,
      };

      break;
    }

    case 'CLEAR_CART': {
      newState = {
        items: [],
        totalCents: 0,
      };
      
      break;
    }
    
    case 'LOAD_FROM_STORAGE': {
      newState = action.cart;
      break;
    }
    
    default:
      return state;
  }
  
  // Save to localStorage for all actions except LOAD_FROM_STORAGE
  if (action.type !== 'LOAD_FROM_STORAGE') {
    saveCartToStorage(newState);
  }
  
  return newState;
}

const CartContext = createContext<CartContextType | undefined>(undefined);

const initialCart: Cart = {
  items: [],
  totalCents: 0,
};

export function CartProvider({ children }: { children: ReactNode }) {
  const [cart, dispatch] = useReducer(cartReducer, initialCart);
  
  // Load cart from localStorage on mount
  useEffect(() => {
    const storedCart = loadCartFromStorage();
    if (storedCart) {
      dispatch({ type: 'LOAD_FROM_STORAGE', cart: storedCart });
    }
  }, []);
  
  const addToCart = (product: ProductDTO, configuration?: CartItemConfiguration) => {
    dispatch({ type: 'ADD_TO_CART', product, configuration });
  };

  const updateItemConfiguration = (
    oldItemId: string,
    product: ProductDTO,
    configuration?: CartItemConfiguration
  ) => {
    dispatch({ type: 'REPLACE_ITEM', oldItemId, product, configuration });
  };
  
  const removeFromCart = (itemId: string) => {
    dispatch({ type: 'REMOVE_FROM_CART', itemId });
  };
  
  const updateQuantity = (itemId: string, quantity: number) => {
    dispatch({ type: 'UPDATE_QUANTITY', itemId, quantity });
  };
  
  const clearCart = () => {
    dispatch({ type: 'CLEAR_CART' });
  };
  
  const getItemQuantity = (productId: string, configuration?: CartItemConfiguration): number => {
    const itemId = generateCartItemId({ id: productId } as ProductDTO, configuration);
    const item = cart.items.find(item => item.id === itemId);
    return item ? item.quantity : 0;
  };
  
  const getTotalProductQuantity = (productId: string): number => {
    return cart.items
      .filter(item => item.product.id === productId)
      .reduce((total, item) => total + item.quantity, 0);
  };
  
  const contextValue: CartContextType = {
    cart,
    addToCart,
    updateItemConfiguration,
    removeFromCart,
    updateQuantity,
    clearCart,
    getItemQuantity,
    getTotalProductQuantity,
  };
  
  return (
    <CartContext.Provider value={contextValue}>
      {children}
    </CartContext.Provider>
  );
}

export function useCart() {
  const context = useContext(CartContext);
  if (context === undefined) {
    throw new Error('useCart must be used within a CartProvider');
  }
  return context;
}
