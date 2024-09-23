import { createContext, useState } from 'react';

const CartContext = createContext();

const CartProvider = ({ children }) => {
  const [cartItems, setCartItems] = useState([]); // Define cartItems state and setCartItems function

  const updateCart = (newCartItems) => {
    setCartItems(newCartItems); // Update cartItems state using setCartItems function
  };

  return (
    <CartContext.Provider value={{ cartItems, updateCart }}>
      {children}
    </CartContext.Provider>
  );
};

export { CartProvider, CartContext };