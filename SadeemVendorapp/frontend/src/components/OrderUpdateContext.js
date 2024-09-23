// OrderUpdateContext.js
import React, { createContext, useContext, useState } from 'react';

const OrderUpdateContext = createContext();

export const OrderUpdateProvider = ({ children }) => {
  const [shouldUpdateOrders, setShouldUpdateOrders] = useState(false);

  return (
    <OrderUpdateContext.Provider value={{ shouldUpdateOrders, setShouldUpdateOrders }}>
      {children}
    </OrderUpdateContext.Provider>
  );
};

export const useOrderUpdate = () => useContext(OrderUpdateContext);
