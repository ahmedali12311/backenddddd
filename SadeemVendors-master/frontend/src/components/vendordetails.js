
import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useParams,  } from 'react-router-dom';
import '../css/vendordetails.css';
import defaultImage from '../css/vendor.jpg';
import { CSSTransition, TransitionGroup } from 'react-transition-group';
import { useOrderUpdate } from './OrderUpdateContext'; // Import the custom hook

function VendorDetails({ onAddToCart }) {
  const { id } = useParams();
  const [searchTerm, setSearchTerm] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage] = useState(12 ); // Adjust this number as needed
  const [vendor, setVendor] = useState(null);
  const [admins, setAdmins] = useState([]);
  const [tables, setTables] = useState([]);
  const [items, setItems] = useState([]);
  const [newAdminEmail, setNewAdminEmail] = useState('');
  const [error, setError] = useState(null);
  const [showSuccessMessage, setShowSuccessMessage] = useState(false);
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [userRole, setUserRole] = useState(null);
  const imageRef = useRef(null);
  const [errorMessage, setErrorMessage] = useState(null);
  const [showAddTableInput, setShowAddTableInput] = useState(false);
  const [newTableName, setNewTableName] = useState('');
  const [setErrorTimeout] = useState(null);
  const [newItemName, setNewItemName] = useState('');
  const [newItemPrice, setNewItemPrice] = useState('');
  const [editingItemId,setEditingItemId] = useState(null);
  const [editItemName, setEditItemName] = useState('');
  const [editItemPrice, setEditItemPrice] = useState('');
  const [editItemImage, setEditItemImage] = useState('');
  const [isEditingTable, setIsEditingTable] = useState(false);
  const [editingTableId, setEditingTableId] = useState(null);
  const [sortOrder,setSortOrder] = useState('created_at'); // Adjust as needed
  const [editTableName, setEditTableName] = useState('');
  const [preview, setPreview] = useState('');
  const [showAddItemInput, setShowAddItemInput] = useState(false);
  const firstItemRef = useRef(null);
  const [isEditingItem, setIsEditingItem] = useState(false);
  const [isTransitioning, setIsTransitioning] = useState(false);
  const [filteredItems, setFilteredItems] = useState([]);
  const [editItemDiscount, setEditItemDiscount] = useState(''); // State for discount
const [totalItemsCount, setTotalItemsCount] = useState(0);
const [newItemDiscount, setNewItemDiscount] = useState(''); // Add this line
const [newItemDiscountDays, setNewItemDiscountDays] = useState('');
const [editItemDiscountDays, setEditItemDiscountDays] = useState('');
const [originalDiscount, setOriginalDiscount] = useState(editItemDiscount); // Add this state
const [isEditingDiscount, setIsEditingDiscount] = useState(false);
const visiblePages = 4; // number of page numbers to display initially
const [itemQuantities, setItemQuantities] = useState({}); // State for item-specific quantities
const [successMessage, setSuccessMessage] = useState('');
const [itemMessages, setItemMessages] = useState({});
const MESSAGE_DURATION = 3000; // Duration in milliseconds
const [itemNameError, setItemNameError] = useState('');
const [itemPriceError, setItemPriceError] = useState('');
const [itemDiscountError, setItemDiscountError] = useState('');
const [itemDiscountDaysError, setItemDiscountDaysError] = useState('');
const [ setIsAddingItem] = useState(false); // To track item addition status
const { shouldUpdateOrders, setShouldUpdateOrders } = useOrderUpdate();
const [newItemQuantity, setNewItemQuantity] = useState(1);
const [editItemQuantity, setEditItemQuantity] = useState(1);
const [itemQuantityError, setItemQuantityError] = useState('');
const [isEditingQuantity, setIsEditingQuantity] = useState(false);

////////////////carts and items////////////////////
const handleAnimationEnd = () => {
  setTimeout(() => {
    setSuccessMessage('');
  }, MESSAGE_DURATION);
};
  // Initialize state to manage quantities for each item


  const handleQuantityChange = (itemId, delta) => {
    setItemQuantities(prevQuantities => ({
      ...prevQuantities,
      [itemId]: Math.max(1, (prevQuantities[itemId] || 1) + delta) // Ensure quantity is at least 1
    }));
  };
  const handleAddToCart = async (item) => {
    const token = localStorage.getItem('token');
    
    // Get the quantity for the specific item, default to 1 if invalid
    let quantity = itemQuantities[item.id];
    if (isNaN(quantity) || quantity < 1) {
      quantity = 1; // Default to 1 if quantity is invalid
      setItemQuantities(prevQuantities => ({
        ...prevQuantities,
        [item.id]: quantity, // Update the state with the default quantity
      }));
    }
  
    try {
      const response = await fetch('http://localhost:8080/cartitems', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          'Authorization': `Bearer ${token}`,
        },
        body: new URLSearchParams({
          item_id: item.id,
          quantity: quantity,
        }).toString(),
      });
  
      if (!response.ok) {
        const errorData = await response.json();
        setItemMessages(prevMessages => ({
          ...prevMessages,
          [item.id]: `Failed to add item to cart: ${errorData.error || 'Unknown error'}`,
        }));
        setTimeout(() => {
          setItemMessages(prevMessages => ({
            ...prevMessages,
            [item.id]: ''
          }));
        }, MESSAGE_DURATION);
        throw new Error(errorData.error || 'Unknown error');
      }
  
      const data = await response.json();
      setItemMessages(prevMessages => ({
        ...prevMessages,
        [item.id]: '', // Clear any existing item-specific messages
      }));
      setSuccessMessage("Item successfully added to cart!");
      setShowSuccessMessage(true);    
      setTimeout(() => {
        setSuccessMessage('');
      }, MESSAGE_DURATION);
  
      onAddToCart(data.cartItems); // Trigger cart refresh
    } catch (error) {
      setItemMessages(prevMessages => ({
        ...prevMessages,
        [item.id]: `Error: ${error.message || 'An unknown error occurred while adding to cart.'}`,
      }));
      setTimeout(() => {
        setItemMessages(prevMessages => ({
          ...prevMessages,
          [item.id]: ''
        }));
      }, MESSAGE_DURATION);
    }
  };
  
useEffect(() => {
  let timeoutId;
  if (showSuccessMessage) {
    timeoutId = setTimeout(() => {
      setShowSuccessMessage(false);
    }, MESSAGE_DURATION);
  }
  return () => clearTimeout(timeoutId);
}, [showSuccessMessage]);

const handleErrorMessageTimeout = () => {
      setErrorMessage(null);
      setError(null); // Clear the error state as well
    };
       // existing state variables...
       const fetchVendorItems = useCallback(async () => {
        const token = localStorage.getItem('token');
        try {
          const response = await fetch(`http://localhost:8080/vendor/${id}/items?page=${currentPage}&page_size=${itemsPerPage}`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });
      
          if (!response.ok) {
            throw new Error('Failed to fetch items');
          }
      
          const data = await response.json();
          
          setItems(data.items || []);
            setFilteredItems(data.items || []); // Initialize filtered items
        } catch (error) {
          console.error("Error fetching items:", error);
          setError('An unexpected error occurred while fetching vendor items');
        }
      }, [id, currentPage, itemsPerPage]);
      const handleImageClick = (e) => {
        const file = e.target.files[0];
        if (file) {
          setEditItemImage(file); // Update editItemImage state with the new file
          const reader = new FileReader();
          reader.onloadend = () => {
            setPreview(reader.result);
          };
          reader.readAsDataURL(file);
        } else {
          setPreview(defaultImage); // Set to default if no file selected
        }
      };
      useEffect(() => {
        fetchVendorItems();
      }, [fetchVendorItems, newItemName]);
    useEffect(() => {
      let sortedItems = [...items];
    
      if (sortOrder === 'created_at') {
        sortedItems.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
      } else if (sortOrder === 'price') {
        sortedItems.sort((a, b) => a.price - b.price);
      } else if (sortOrder === 'name') {
        sortedItems.sort((a, b) => a.name.localeCompare(b.name));
      }
    
      setFilteredItems(sortedItems);
    }, [sortOrder, items]);


  
const handleNextPage = () => {
  if (currentPage < totalPages) {
      setIsTransitioning(true);
      setCurrentPage(prevPage => prevPage + 1);
  }
};

const handlePreviousPage = () => {
  if (currentPage > 1) {
      setIsTransitioning(true);
      setCurrentPage(prevPage => prevPage - 1);
  }
};

useEffect(() => {
  if (isTransitioning) {
      console.log("Transitioning...");
      fetchVendorItems().then(() => {
          setIsTransitioning(false);
          console.log("Fetching items and ending transition.");

          // Scroll after the items are rendered
          requestAnimationFrame(() => {
              if (firstItemRef.current) {
                  const offsetPosition = firstItemRef.current.getBoundingClientRect().top + window.scrollY - 250; // Adjust the offset
                  window.scrollTo({
                      top: offsetPosition,
                      behavior: 'smooth'
                  });
              }
          });
      });
  }
}, [isTransitioning, fetchVendorItems]);
const fetchUserOrders = useCallback(async () => {
  try {
    const token = localStorage.getItem('token');
    const response = await fetch('http://localhost:8080/orders', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      const errorData = await response.json();
      const errorMessage = errorData.error || 'Failed to fetch orders';
      setErrorMessage(errorMessage);
      return [];
    }

    const data = await response.json();
    setOrders(data.orders || []);
  } catch (error) {
    setErrorMessage('An unexpected error occurred while fetching user orders');
    setTimeout(() => {
      setErrorMessage(null);
    }, 3000);
  }
}, []);
useEffect(() => {
  fetchUserOrders();
}, [fetchUserOrders, shouldUpdateOrders]); // Trigger fetch when shouldUpdateOrders changes


useEffect(() => {
  if (shouldUpdateOrders) {
    fetchUserOrders();
    setShouldUpdateOrders(false); // Reset the update flag
  }
}, [shouldUpdateOrders, fetchUserOrders, setShouldUpdateOrders]);

    const handleSearchChange = (event) => {
      const searchValue = event.target.value;
      setSearchTerm(searchValue);
  
      if (searchValue) {
          const results = items.filter(item => 
              item.name.toLowerCase().includes(searchValue.toLowerCase())
          );
          setFilteredItems(results);
      } else {
          setFilteredItems(items); // Reset to all items when search term is empty
      }
  };
  const fetchVendorDetails = useCallback(async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendors/${id}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        const errorMessage = errorData.error || 'An error occurred while fetching vendor details';
        setError(errorMessage);
        return;
      }

      const data = await response.json();
      if (data && data.vendor) {
        setVendor(data.vendor);
      } else {
        setError('Vendor data is undefined or null');
      }
    } catch (error) {
      setError('An unexpected error occurred while fetching vendor details');
    } finally {
      setLoading(false);
    }
  }, [id]);
  const [currentUserId, setCurrentUserId] = useState(null); // Add this line
  const fetchUserRole = useCallback(async () => {
    try {
      const token = localStorage.getItem('token');
      if (token) {
        const response = await fetch('http://localhost:8080/me', {
          headers: {
            Authorization:`Bearer ${token}`,
          },
        });
  
        if (!response.ok) {
          const errorData = await response.json();
          const errorMessage = errorData.error || 'Failed to fetch user role';
          setError(errorMessage);
          return;
        }
  
        const data = await response.json();
        setUserRole(data.me.user_role);
        setCurrentUserId(data.me.user_info.id); // Update the currentUserId state correctly
      }
    } catch (error) {
      setError('An unexpected error occurred while fetching user role');
    }
  }, []);

  const totalPages = Math.ceil(totalItemsCount / itemsPerPage);
const fetchTotalItemsCount = useCallback(async () => {
  try {
    const token = localStorage.getItem('token');
    const response = await fetch(`http://localhost:8080/vendor/${id}/itemscount`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      const errorData = await response.json();
      const errorMessage = errorData.error || 'Failed to fetch total items count';
      setError(errorMessage);
      return;
    }

    const data = await response.json();
    setTotalItemsCount(data.totalCount); // Update the state with the totalCount property
  } catch (error) {
    setError('An unexpected error occurred while fetching total items count');
  }
}, [id]);
  const fetchVendorAdmins = useCallback(async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendors/${id}/admins`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        const errorMessage = errorData.error || 'Failed to fetch vendor admins';
        setError(errorMessage);
        return;
      }

      const data = await response.json();
      setAdmins(data.vendor_admin || []);
    } catch (error) {
      setError('An unexpected error occurred while fetching vendor admins');
    }
  }, [id]);

  const fetchVendorTables = useCallback(async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendor/${id}/tables`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      if (!response.ok) {
        const errorData = await response.json();
        const errorMessage = errorData.error || 'Failed to fetch vendor tables';
        setError(errorMessage);
        return;
      }

      const data = await response.json();
      console.log(data)

      setTables(data.tables || []);
    } catch (error) {
      setError('An unexpected error occurred while fetching vendor tables');
    }
  }, [id]);
  
  const fetchUserTables = useCallback(async () => {
    try {
      const token = localStorage.getItem("token");
      const response = await fetch("http://localhost:8080/usertable", {
        headers: {
          Authorization:`Bearer ${token}`,
        },
      });
  
      if (!response.ok) {
        const errorData = await response.json();
        const errorMessage = errorData.error;
        if (errorMessage === "User already have a table") {
          setErrorMessage("You already have a table!"); // Updated message
        } else if (errorMessage !== "User has no table") {
          setError(errorMessage);
        }
        return []; // Return an empty array on error
      }
  
      const data = await response.json();
      return Array.isArray(data.tables) ? data.tables : []; // Ensure this always returns an array
      
    } catch (error) {
      setErrorMessage('An unexpected error occurred while fetching user tables');
      setErrorTimeout(
        setTimeout(() => {
          setErrorMessage(null);
        }, 3000)
      );
    }
}, [setErrorTimeout]);
useEffect(() => {
    if (errorMessage) {
      const timeoutId = setTimeout(handleErrorMessageTimeout, 3000); // 3000ms = 3 seconds
      return () => clearTimeout(timeoutId);
    }
  
    if (error) {
      const timeoutId = setTimeout(handleErrorMessageTimeout, 3000); // 3000ms = 3 seconds
      return () => clearTimeout(timeoutId);
    }
  }, [error, errorMessage]);
  useEffect(() => {
    const fetchData = async () => {
      try {
        await fetchVendorDetails();
        await fetchUserRole();
        await fetchVendorAdmins();
        await fetchVendorTables();
        await fetchTotalItemsCount();
        await fetchVendorItems();
        await fetchUserTables();
        await fetchUserOrders();
        setLoading(false); // Set loading to false after fetching data
      } catch (error) {
        console.error('Error fetching data:', error);
        setError('An error occurred while fetching data.');
        setLoading(false); // Set loading to false in case of error
      }
    };
    fetchData();
  }, [fetchVendorDetails, fetchUserRole, fetchVendorAdmins, fetchVendorTables, fetchVendorItems, fetchUserTables, fetchUserOrders, fetchTotalItemsCount]);

useEffect(() => {
  const totalPages = Math.ceil(totalItemsCount / itemsPerPage);
  console.log(`Total pages: ${totalPages}`);
}, [totalItemsCount, itemsPerPage]);
useEffect(() => {
  const totalPages = Math.ceil(totalItemsCount / itemsPerPage);
  console.log(`Total pages: ${totalPages}`);
}, [totalItemsCount, itemsPerPage]);
const handleImageHover = () => {
  if (imageRef.current) {
      imageRef.current.style.cursor = 'pointer';
  }
};

const handleImageLeave = () => {
  if (imageRef.current) {
      imageRef.current.style.cursor = 'default';
  }
};

const handleImageError = (e) => {
  e.target.src = defaultImage;
};
  const handleAddAdmin = async () => {
    if (!newAdminEmail) return;

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendors/${id}/admins`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          Authorization:`Bearer ${token}`,
        },
        body: new URLSearchParams({ Email: newAdminEmail }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        const errorMessage = errorData.error || 'Failed to add admin';
        setError(errorMessage);
        return;
      }

      setNewAdminEmail('');
      fetchVendorAdmins(); // Refresh admin list
    } catch (error) {
      setError('An unexpected error occurred while adding admin');
    }
  };

  const handleRemoveAdmin = async (adminId) => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendors/${id}/admins/${adminId}`, {
        method: 'DELETE',
        headers: {
          Authorization:`Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        const errorMessage = errorData.error || 'Failed to remove admin';
        setError(errorMessage);
        return;
      }

      // Remove the admin from the list
      setAdmins((prevAdmins) => prevAdmins.filter((admin) => admin.user_id !== adminId));
    } catch (error) {
      setError('An unexpected error occurred while removing admin');
    }
  };
  const handleUpdateQuantity = async () => {
    console.log('handleUpdateQuantity called');
    setItemQuantityError('');
    if (editItemQuantity <= 0) {
      setItemQuantityError('Quantity must be greater than 0.');
      return;
    }
  
    try {
      const token = localStorage.getItem('token');
      console.log('Token:', token);
  
      const body = new URLSearchParams({ quantity: editItemQuantity }).toString();
      console.log('Request Body:', body);
  
      const response = await fetch(`http://localhost:8080/vendor/${id}/items/${editingItemId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          Authorization: `Bearer ${token}`,
        },
        body: body,
      });
  
      console.log('Response Status:', response.status);
  
      if (!response.ok) {
        const errorData = await response.json();
        console.log('Error Data:', errorData);
        setItemQuantityError(errorData.error?.quantity || '');
        return;
      }
  
      // Successfully updated quantity, reset fields and close modal
      setIsEditingQuantity(false);
      setEditItemQuantity(1);
      fetchVendorItems(); // Refresh item list
    } catch (error) {
      console.error('Error updating quantity:', error)
    }
  }  
  
  const handleRemoveTable = async (tableId) => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendor/${id}/tables/${tableId}`, {
        method: 'DELETE',
        headers: {
          Authorization:`Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        const errorMessage = errorData.error || 'Failed to remove table';
        setError(errorMessage);
        return;
      }

      // Remove the table from the list
      setTables((prevTables) => prevTables.filter((table) => table.id !== tableId));
    } catch (error) {
      setError('An unexpected error occurred while removing table');
    }
  };



  if (loading) {
    return <div  className="spinner"></div>;
  }
  const handleAddTable = async () => {
    if (!newTableName) return;

    try {
        const token = localStorage.getItem('token');
        const response = await fetch(`http://localhost:8080/vendor/${id}/tables`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
                Authorization: `Bearer ${token}`,
            },
            body: new URLSearchParams({ name: newTableName }),
        });

        if (!response.ok) {
            const errorData = await response.json();
            const errorMessage = errorData.error || 'Failed to add table';
            setError(errorMessage);
            return;
        }

        setNewTableName('');
        setShowAddTableInput(false);
        fetchVendorTables(); // Refresh table list
    } catch (error) {
        setError('An unexpected error occurred while adding the table');
    }
};

const handleUpdateDiscount = async () => {
 
  setItemDiscountError('');
  setItemDiscountDaysError('');
  const body = new URLSearchParams();

  // If the discount is set to 0
  if (editItemDiscount === 0) {
      body.append('discount', 0);
      setEditItemDiscount(0); // Reset state
      setEditItemDiscountDays(0); // Reset expiration days to 0
      body.append('discount_days', 0); // Also set discount_days to 0
  } else if (editItemDiscount !== originalDiscount) {
      body.append('discount', editItemDiscount);
  }

  // Only add expiration days if they are provided and discount is not 0
  if (editItemDiscountDays > 0) {
      body.append('discount_days', editItemDiscountDays);
  }

  // Prevent submission if the discount has changed and expiration days are required
  if (editItemDiscount !== originalDiscount && editItemDiscount > 0 && !editItemDiscountDays) {
      alert("Please enter the discount expiration days.");
      return; // Prevent submission if expiration days are required
  }

  // Prevent submission if no changes are made
  if (body.toString() === '') {
      setIsEditingDiscount(false);
      return;
  }

  try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendor/${id}/items/${editingItemId}`, {
          method: 'PUT',
          headers: {
              'Content-Type': 'application/x-www-form-urlencoded',
              Authorization: `Bearer ${token}`,
          },
          body: body,
      });
      const data = await response.json();

      if (!response.ok) {
        setItemDiscountError(data.error.discount || '');
        setItemDiscountDaysError(data.error.discount_expiry || '');
        return;
    }

      // Successfully updated
      setIsEditingDiscount(false);
      setEditItemDiscount(0); // Reset to 0
      setEditItemDiscountDays(0); // Reset to 0
      fetchVendorItems(); // Optionally refresh the item list
  } catch (error) {
      setError('An unexpected error occurred while updating the discount');
  }
};
const handleEditItem = (item) => {
  setEditingItemId(item.id);
  setEditItemName(item.name);
  setEditItemPrice(item.price);
  setEditItemImage(item.img);
  setOriginalDiscount(item.discount);
  setEditItemQuantity(item.quantity); 

  setEditItemDiscount(0); // Reset discount to 0
  setEditItemDiscountDays(0); // Reset discount days to 0
  openEditItemForm();

};
const handleUpdateItem = async () => {
  setItemNameError('');
  setItemPriceError('');
  setItemQuantityError('');
  setItemDiscountError('');
  setItemDiscountDaysError('');
  
  // Validate inputs
  if (!editItemName || !editItemPrice) {
      if (!editItemName) {
          setItemNameError('Item name is required.');
      }
      if (!editItemPrice) {
          setItemPriceError('Price is required.');
      }
      return;
  }
  
  // Client-side validation
  if (editItemName.length > 20) {
      setItemNameError('Item name must be less than 20 characters.');
      return;
  }
  if (editItemPrice <= 0) {
      setItemPriceError('Price must be greater than 0.');
      return;
  }
  if (editItemQuantity <= 0) {
      setItemQuantityError('Quantity must be greater than 0.');
      return;
  }
  
  // Prepare form data
  const body = new FormData();
  if (editItemImage) {
    body.append('img', editItemImage);
}

  body.append('name', editItemName);
  body.append('price', editItemPrice);
  body.append('quantity', editItemQuantity);
  if (editItemDiscount) {
      body.append('discount', editItemDiscount);
      body.append('discount_days', editItemDiscountDays);
  }
  
  // Send the request
  try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendor/${id}/items/${editingItemId}`, {
          method: 'PUT',
          headers: {
              Authorization: `Bearer ${token}`,
          },
          body: body,
      });
      
      if (!response.ok) {
          const errorData = await response.json();
          setItemNameError(errorData.error?.name || '');
          setItemPriceError(errorData.error?.price || '');
          setItemQuantityError(errorData.error?.quantity || '');
          setItemDiscountError(errorData.error?.discount || '');
          setItemDiscountDaysError(errorData.error?.discountDays || '');
          return;
      }
      
      // Successfully updated item, reset fields and close modal
      setIsEditingItem(false);
      setEditItemName('');
      setEditItemPrice('');
      setEditItemQuantity('');
      setEditItemImage('');
      setEditItemDiscount('');
      setEditItemDiscountDays('');
      fetchVendorItems(); // Refresh item list
  } catch (error) {
      console.error('Error updating item:', error);
  }
};

const resetErrors = () => {
  
  setItemNameError('');
  setItemPriceError('');
  setItemDiscountError('');
  setItemDiscountDaysError('');
};
const openEditItemForm = () => {
  resetErrors();
  setIsEditingItem(true);
};

const openEditDiscountForm = () => {
  resetErrors();
  setIsEditingDiscount(true);
};
const handleAddItem = async () => {
  setItemNameError('');
  setItemPriceError('');
  setItemDiscountError('');
  setItemDiscountDaysError('');

  const formData = new FormData();
if (newItemQuantity <= 0) {
  setItemQuantityError('Quantity must be greater than 0.');
  return;
}

formData.append('quantity', newItemQuantity);
  if (imageRef.current?.files[0]) {
      formData.append('img', imageRef.current.files[0]);
  }

  // Validation checks
  if (!newItemName || !newItemPrice) {
      if (!newItemName) {
          setItemNameError('Item name is required.');
      }
      if (!newItemPrice) {
          setItemPriceError('Price is required.');
      }
      return;
  }

  // Client-side validation
  if (newItemName.length > 20) {
      setItemNameError('Item name must be less than 20 characters.');
      return;
  }

  if (newItemPrice <= 0) {
      setItemPriceError('Price must be greater than 0.');
      return;
  }

  if (newItemDiscount < 0) {
      setItemDiscountError('Discount cannot be negative.');
      return;
  }

  if (newItemDiscount > newItemPrice) {
      setItemDiscountError('Discount cannot be greater than the price.');
      return;
  }

  try {
      const token = localStorage.getItem('token');

      formData.append('name', newItemName);
      formData.append('price', newItemPrice);
      formData.append('discount', newItemDiscount);
      formData.append('discount_days', newItemDiscountDays);

      const response = await fetch(`http://localhost:8080/vendor/${id}/items`, {
          method: 'POST',
          headers: {
              Authorization: `Bearer ${token}`,
          },
          body: formData,
      });

  
      if (!response.ok) {
        const errorData = await response.json();
        console.error('Backend errors:', errorData); // Log the error response for debugging
        
        // Clear existing frontend errors
        setItemNameError('');
        setItemPriceError('');
        setItemDiscountError('');
        setItemDiscountDaysError('');

        // Set specific backend errors
        if (errorData.error) {
            if (errorData.error.name) setItemNameError(errorData.error.name);
            if (errorData.error.price) setItemPriceError(errorData.error.price);
            if (errorData.error.discount) setItemDiscountError(errorData.error.discount);
            if (errorData.error.discount_days) setItemDiscountDaysError(errorData.error.discount_days);
        }

        setPreview(null); // Reset preview on error
        return;
    }
      // Reset fields and close the modal only if submission is successful
      setShowAddItemInput(false); // Close modal only on successful request

      setNewItemName('');
      setNewItemPrice('');
      setNewItemDiscount('');
      setNewItemDiscountDays('');
      setPreview(null); // Reset preview after successful submission
      setIsAddingItem(true); // Mark adding as successful
      setItemNameError('');
      setItemPriceError('');
      setItemDiscountError('');
      setItemDiscountDaysError('');
      fetchVendorItems(); // Refresh items

  } catch (error) {
      console.error('Error adding item:', error);
  }
};


  const handleOccupyTable = async (tableId) => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendor/${id}/tables/${tableId}/needs-service`, {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
  
      if (!response.ok) {
        const errorData = await response.json();
        const errorMessage = errorData.error || 'Failed to occupy the table';
        setError(errorMessage);
        return;
      }
  
      // Update the local state to reflect that the table is now occupied
      setTables((prevTables) =>
        prevTables.map((table) =>
          table.id === tableId ? { ...table, is_available: false, is_needs_service: true, customer_id: currentUserId } : table
        )
      );
  
      console.log('Table occupied successfully');
    } catch (error) {
      setError('An unexpected error occurred while occupying the table');
    }
  };
// New useEffect for handling assigned tables
const handleFreeTable = async (tableId) => {
  const table = tables.find((t) => t.id === tableId);
  if (!table || table.customer_id !== currentUserId) {
      setError('You can only free the table you have assigned.');
      return;
  }

  try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendor/${id}/tables/${tableId}/freetable`, {
          method: 'PUT',
          headers: {
              Authorization: `Bearer ${token}`,
          },
      });

      if (!response.ok) {
          const errorData = await response.json();
          const errorMessage = errorData.error || 'Failed to free the table';
          setError(errorMessage);
          return;
      }

      // Update the table state to reflect that it is now available
      setTables((prevTables) =>
          prevTables.map((table) =>
              table.id === tableId ? { ...table, is_available: true, is_needs_service: false, customer_id: null } : table
          )
      );

      // Update the orders state to null if the table being freed matches the table associated with the orders
      if (orders && orders.some((order) => order.table_id === tableId)) {
        setOrders(null);
      }
  } catch (error) {
      setError('An unexpected error occurred while freeing the table');
  }
};
const handleRemoveDiscount = async (itemId) => {
  const body = new URLSearchParams();
  body.append('discount', 0);
  body.append('discount_days', 0);

  try {
    const token = localStorage.getItem('token');
    const response = await fetch(`http://localhost:8080/vendor/${id}/items/${itemId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        Authorization: `Bearer ${token}`,
      },
      body: body,
    });

    if (!response.ok) {
      const errorData = await response.json();
      setError(errorData.error || 'Failed to remove discount');
      return;
    }

    // Update the specific item that had its discount removed
    const updatedItem = { ...items.find((item) => item.id === itemId), discount: 0, discountDays: 0 };
    setItems((prevItems) => prevItems.map((item) => (item.id === itemId ? updatedItem : item)));
  } catch (error) {
    setError('An unexpected error occurred while removing the discount');
  }
};

const handleRemoveItem = async (itemId) => {
  try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendor/${id}/items/${itemId}`, {
          method: 'DELETE',
          headers: {
              Authorization: `Bearer ${token}`,
          },
      });

      if (!response.ok) {
          const errorData = await response.json();
          const errorMessage = errorData.error || 'Failed to remove item';
          setError(errorMessage);
          return;
      }

      // Remove the item from local state
      setItems((prevItems) => prevItems.filter((item) => item.id !== itemId));
  } catch (error) {
      setError('An unexpected error occurred while removing the item');
  }
};

const startPage = Math.max(1, currentPage - visiblePages + 1);
const endPage = Math.min(totalPages, currentPage + visiblePages - 1);

const pages = [];
for (let i = startPage; i <= endPage; i++) {
  pages.push(i);
}
const handleUpdateNeedsService = async (tableId, isNeedsService) => {
  try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendor/${id}/table/${tableId}`, {
          method: 'PUT',
          headers: {
              'Content-Type': 'application/x-www-form-urlencoded',
              Authorization:`Bearer ${token}`,
          },
          body: new URLSearchParams({ is_needs_service: isNeedsService }),
      });

      if (!response.ok) {
          const errorData = await response.json();
          const errorMessage = errorData.error || 'Failed to update needs service status';
          setError(errorMessage);
          return;
      }

      // Update the local state
      setTables((prevTables) =>
          prevTables.map((table) =>
              table.id === tableId ? { ...table, is_needs_service: isNeedsService } : table
          )
      );
  } catch (error) {
      setError('An unexpected error occurred while updating the needs service status');
  }
};

  const handleVendorFreeTable = async (tableId) => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendor/${id}/freetable/${tableId}`, {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
  
      if (!response.ok) {
        const errorData = await response.json();
        const errorMessage = errorData.error || 'Failed to free the table';
        console.error(errorMessage);
        setError(errorMessage);
        return;
      }
  
      console.log('Table freed successfully');
      // Update the local state to reflect that the table is now available
      setTables((prevTables) =>
        prevTables.map((table) =>
          table.id === tableId ? { ...table, is_available: true, is_needs_service: false, customer_id: null } : table
        )
      );
    } catch (error) {
      console.error('An unexpected error occurred while freeing the table', error);
      setError('An unexpected error occurred while freeing the table');
    }
  };
  const isCurrentUserAdmin = admins.some(admin => admin.user_id === currentUserId);
  const handleEditTable = (table) => {
    setEditingTableId(table.id);
    setEditTableName(table.name);
    setIsEditingTable(true);
};
const handleSortOrderChange = (newSortOrder) => {
  setSortOrder(newSortOrder);
};
const handleUpdateTable = async () => {
  if (editTableName === undefined) return;

  try {
    const token = localStorage.getItem('token');
    const response = await fetch(`http://localhost:8080/vendor/${id}/table/${editingTableId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        Authorization: `Bearer ${token}`,
      },
      body: new URLSearchParams({ name: editTableName }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      const errorMessage = errorData.error || 'Failed to update table';
      setError(errorMessage);
      return;
    }

    // Refresh the table list
    fetchVendorTables();
    setIsEditingTable(false);
    setEditTableName('');
  } catch (error) {
    setError('An unexpected error occurred while updating the table');
  }
};

  return (

<div className="page-container">

   <div className="vendor-details-container">
    <div className="vendor-background">
      <img
        src={vendor?.img || defaultImage}
        alt={vendor?.name || 'Vendor'}
        className="vendor-background-image"
        ref={imageRef}
        onMouseOver={handleImageHover}
        onMouseOut={handleImageLeave}
        onError={handleImageError}
      />
      <div className="vendor-info">
        <h1 className="vendor-name">{vendor?.name}</h1>
        <p className="vendor-description">{vendor?.description || 'No description available'}</p>
        <button className="items-button" onClick={() => document.getElementById('items-section').scrollIntoView()}>
          View Items
        </button>
      </div>
    </div>
 {/* Tables Section */}
<div className="tables-section card">
{(userRole === '1' || (userRole === '2' && isCurrentUserAdmin)) && (
    <div className="add-buttons">
      <button onClick={() => setShowAddTableInput(true)}>Add Table</button>
    </div>
  )}

  {showAddTableInput && (
    <div className="edit-table-modal">
      <div className="modal-overlay" onClick={() => setShowAddTableInput(false)}></div>
      <div className="edit-table-form">
        <h3>Add Table</h3>
        <form onSubmit={(e) => { e.preventDefault(); handleAddTable(); }}>
          <div className="form-group">
            <label htmlFor="new-table-name">Table Name</label>
            <input
              type="text"
              id="new-table-name"
              value={newTableName}
              onChange={(e) => setNewTableName(e.target.value)}
            />
          </div>
          <button type="submit">Add Table</button>
          <button type="button" onClick={() => setShowAddTableInput(false)}>Cancel</button>
        </form>
      </div>
    </div>
  )}

  <h2 className="section-title">Tables</h2>
  <div className="tables-list">
    {tables.map((table) => (
      <div key={table.id} className="table-card">
        <h3 className="table-name">{table.name}</h3>
        <p>Available: {table.is_available ? 'Yes' : 'No'}</p>
        <p>Needs Service: {table.is_needs_service ? 'Yes' : 'No'}</p>

        <div className="table-actions">
          {table.customer_id === currentUserId && !table.is_available && (
            <button onClick={() => handleFreeTable(table.id)}>Free Table</button>
          )}

          <div className="action-row">
            {table.is_available && (table.customer_id !== currentUserId) && userRole === '3' && (
              <button onClick={() => handleOccupyTable(table.id)} className="occupy-button">Occupy Table</button>
            )}
          </div>

          <div className="service-actions">
            {!table.is_available && (userRole === '1' || (userRole === '2' && isCurrentUserAdmin)) && (
              <>
                <button onClick={() => handleUpdateNeedsService(table.id, !table.is_needs_service)}>
                  {table.is_needs_service ? 'Mark as No Needs Service' : 'Mark as Needs Service'}
                </button>
                <button onClick={() => handleVendorFreeTable(table.id)}>Vendor Free Table</button>
              </>
            )}
          </div>
        </div>

        {table.customer_id === currentUserId && !table.is_available && (
      <div className="orders-list">
      {orders ? (
        orders
          .filter(order => order.table_id === table.id) // Filter orders for the specific table
          .map((order) => (
            <div key={order.id} className="order-card">
              <h3 className="order-id">Order ID: {order.id}</h3>
              {order.item_names.map((itemName, index) => (
                <p key={index}>
                  {itemName}: {order.item_quantities[index]}
                </p>
              ))}
              <p>Total Price: {order.total_order_cost}</p>
              <p>Status: {order.status}</p>
            </div>
          ))
      ) : (
        <p>No orders available.</p>
      )}
    </div>
        )}

        <div className="edit-remove-row">
          {(userRole === '1' || (userRole === '2' && isCurrentUserAdmin)) && (
            <>
              <button onClick={() => handleEditTable(table)}>Edit</button>
              <button onClick={() => handleRemoveTable(table.id)}>Remove</button>
            </>
          )}
        </div>
      </div>
    ))}
  </div>


{/* Full-Screen Edit Form for Tables */}
{isEditingTable && (
    <div className="edit-table-modal">
        <div className="modal-overlay" onClick={() => setIsEditingTable(false)}></div>
        <div className="edit-table-form">
            <h3>Edit Table</h3>
            <form onSubmit={(e) => { e.preventDefault(); handleUpdateTable(); }}>
                <div className="form-group">
                    <label htmlFor="table-name">Table Name</label>
                    <input
                        type="text"
                        id="table-name"
                        value={editTableName}
                        onChange={(e) => setEditTableName(e.target.value)}
                    />
                </div>
                <button type="submit">Save Changes</button>
                <button type="button" onClick={() => setIsEditingTable(false)}>Cancel</button>
            </form>
        </div>
    </div>
)}

{error && error.includes('table') && (
    <div className="error-message">{error}</div>
  )}
  
{/* Items Section */}

<div className="sort-selection">
    <select value={sortOrder} onChange={(e) => handleSortOrderChange(e.target.value)}>
        <option value="created_at">Created at</option>
        <option value="price">Price</option>
        <option value="name">Name</option>
    </select>
</div>

{/* Add Item Modal */}

{showAddItemInput && (
    <div className="edit-item-modal">
        <div className="modal-overlay" onClick={() => setShowAddItemInput(false)}></div>
        <div className="edit-item-form">
            <h3>Add Item</h3>

            <form onSubmit={(e) => {
                e.preventDefault();
                handleAddItem();
            }}>
                <div className="form-group">
                    <label htmlFor="new-item-name">Item Name</label>
                    <input
                        type="text"
                        id="new-item-name"
                        value={newItemName}
                        onChange={(e) => setNewItemName(e.target.value)}
                    />
                    {itemNameError && <div className="error-message">{itemNameError}</div>}
                </div>

                <div className="form-group">
                    <label htmlFor="new-item-price">Price</label>
                    <input
                        type="number"
                        id="new-item-price"
                        value={newItemPrice}
                        onChange={(e) => setNewItemPrice(e.target.value)}
                    />
                    {itemPriceError && <div className="error-message">{itemPriceError}</div>}
                </div>
                <div className="form-group">
    <label htmlFor="newItemQuantity">Quantity</label>
    <input
        type="number"
        id="newItemQuantity"
        value={newItemQuantity}
        onChange={(e) => setNewItemQuantity(e.target.value)}
    />
    {itemQuantityError && <div className="error-message">{itemQuantityError}</div>}
</div>
                <div className="form-group">
                    <label htmlFor="newItemDiscount">Discount</label>
                    <input
                        type="number"
                        id="newItemDiscount"
                        value={newItemDiscount}
                        onChange={(e) => {
                            const discountValue = e.target.value;
                            setNewItemDiscount(discountValue);
                            if (discountValue === '') {
                                setNewItemDiscountDays(''); // Reset discount days if discount is removed
                            }
                        }}
                    />
                    {itemDiscountError && <div className="error-message">{itemDiscountError}</div>}
                </div>

                {newItemDiscount > 0 && (
                    <div className="form-group">
                        <label htmlFor="newItemDiscountDays">Discount Expiration (days)</label>
                        <input
                            type="number"
                            id="newItemDiscountDays"
                            value={newItemDiscountDays}
                            onChange={(e) => setNewItemDiscountDays(e.target.value)}
                        />
                        {itemDiscountDaysError && <div className="error-message">{itemDiscountDaysError}</div>}
                    </div>
                )}

                <div className="form-group">
                    <label htmlFor="new-item-image">Image</label>
                    <input
                        type="file"
                        ref={imageRef}
                        onChange={handleImageClick}
                        style={{ display: 'none' }}
                    />
                    <div className="profile-image-container" onClick={() => imageRef.current?.click()}>
                        <img
                            src={preview || defaultImage}
                            alt="Preview"
                            className="profile-image"
                            onError={handleImageError}
                        />
                    </div>
                </div>

                <button type="submit">ADD</button>
                <button type="button" onClick={() => setShowAddItemInput(false)}>Cancel</button>
            </form>
        </div>
    </div>
)}


{/* Search Input */}
<div className="custom-search-container">
    <input
        type="text"
        value={searchTerm}
        onChange={handleSearchChange}
        placeholder="Search items..."
        className="custom-search-input"
    />
</div>

{/* Items List */}
{(userRole === '1' || (userRole === '2' && isCurrentUserAdmin)) && (
    <div className="add-buttons">
        <button onClick={() => setShowAddItemInput(true)}>Add Item</button>
    </div>
)}
<div className="items-section card items-transition" id="items-section">
    <h2 className="section-title">Items</h2>
    <TransitionGroup className="items-list">
    {filteredItems.map((item, index) => (
  <CSSTransition
    key={item.id}
    classNames="items-transition"
    timeout={500}
  >
    <div className="item-card" ref={index === 0 ? firstItemRef : null}>
      <div className="item-image-container">
        <img
          src={item.img || defaultImage}
          alt={item.name}
          className="item-image"
          onError={handleImageError}
        />
      </div>
      <h3 className="item-name">{item.name}</h3>
      
      {item.discount > 0 ? (
        <>
          <p className="item-price original-price">Price: ${item.price.toFixed(2)}</p>
          <p className="discounted-price">Discounted Price: ${item.discount.toFixed(2)}</p>
        </>
      ) : (
        <p className="item-price">Price: ${item.price.toFixed(2)}</p>
      
      )}
 <p className="item-quantity">Quantity: {item.quantity}</p>
 {item.quantity === 0 ? (
  <p className="out-of-stock">Out of Stock</p>
) : (
  userRole === '3' && (
    <div className="quantity-input">
    <button onClick={() => handleQuantityChange(item.id, -1)}>-</button>
    <span>{itemQuantities[item.id] || 1}</span>
    <button onClick={() => handleQuantityChange(item.id, 1)}>+</button>
    <div>
      <button onClick={() => handleAddToCart(item)}>Add to cart</button>
    </div>
  </div>
  )
)}
      <div className="edit-remove-row">
        
        {(userRole === '1' || (userRole === '2' && isCurrentUserAdmin)) && (
          <div className="item-actions">
            <button onClick={() => handleEditItem(item)}>Edit</button>
            <button onClick={() => handleRemoveItem(item.id)}>Remove</button>
          </div>
        )}
                  {(userRole === '1' || (userRole === '2' && isCurrentUserAdmin)) && (

<button onClick={() => {
setEditItemQuantity(item.quantity);
setEditingItemId(item.id);
setIsEditingQuantity(true);
}}>
Edit Quantity
            
</button>
            )}
        <div className="discount-actions">
          {(userRole === '1' || (userRole === '2' && isCurrentUserAdmin)) && (
            <>
              <button onClick={() => {
                setEditItemDiscount(item.discount);
                setEditItemDiscountDays(item.discountDays || '');
                setOriginalDiscount(item.discount);
                setEditingItemId(item.id);
                openEditDiscountForm();                
              }}>
                Add Discount
              </button>
              <button type="button" onClick={() => handleRemoveDiscount(item.id)}>
                Remove Discount
              </button>
            </>
          )}
        </div>
      
      </div>
      {itemMessages[item.id] && (
        <div className="item-message">
          {itemMessages[item.id]}
        </div>
      )}
    </div>
  </CSSTransition>
))}

<div
  className={`success-message ${!successMessage ? 'success-message-hidden' : ''}`}
  onAnimationEnd={handleAnimationEnd}
>
  {successMessage}
</div>
    </TransitionGroup>
</div>

{error && error.includes('item') && (
    <div className="error-message">{error}</div>
)}
{/* Full-Screen Edit Form */}
{isEditingItem && (
    <div className="edit-item-modal">
        <div className="modal-overlay" onClick={() => setIsEditingItem(false)}></div>
        <div className="edit-item-form">
            <h3>Edit Item</h3>
            <form onSubmit={(e) => { e.preventDefault(); handleUpdateItem(); }}>
                <div className="form-group">
                    <label htmlFor="name">Name</label>
                    <input
                        type="text"
                        id="name"
                        value={editItemName}
                        onChange={(e) => setEditItemName(e.target.value)}
                    />
                    {itemNameError && <div className="error-message">{itemNameError}</div>}
                </div>
                <div className="form-group">
                    <label htmlFor="price">Price</label>
                    <input
                        type="number"
                        id="price"
                        value={editItemPrice}
                        onChange={(e) => setEditItemPrice(e.target.value)}
                    />
                    {itemPriceError && <div className="error-message">{itemPriceError}</div>}
                </div>
                <div className="form-group">
                    <label htmlFor="editItemQuantity">Quantity</label>
                    <input
                        type="number"
                        id="editItemQuantity"
                        value={editItemQuantity}
                        onChange={(e) => setEditItemQuantity(e.target.value)}
                    />
                    {itemQuantityError && <div className="error-message">{itemQuantityError}</div>}
                </div>
                <div className="form-group">
                    <label htmlFor="discount">Discount</label>
                    <input
                        type="number"
                        id="discount"
                        value={editItemDiscount}
                        onChange={(e) => setEditItemDiscount(e.target.value)}
                    />
                    {itemDiscountError && <div className="error-message">{itemDiscountError}</div>}
                </div>
                {editItemDiscount > 0 && (
                    <div className="form-group">
                        <label htmlFor="discountDays">Discount Expiration (days)</label>
                        <input
                            type="number"
                            id="discountDays"
                            value={editItemDiscountDays}
                            onChange={(e) => setEditItemDiscountDays(e.target.value)}
                        />
                        {itemDiscountDaysError && <div className="error-message">{itemDiscountDaysError}</div>}
                    </div>
                )}
                <div className="form-group">
                    <label htmlFor="image">Image</label>
                    <input
                        type="file"
                        ref={imageRef}
                        onChange={handleImageClick}
                        style={{ display: 'none' }}
                    />
                    <div className="profile-image-container" onClick={() => imageRef.current?.click()}>
                        <img
                            src={preview || editItemImage || defaultImage}
                            alt="Preview"
                            className="profile-image"
                            onError={handleImageError}
                        />
                    </div>
                </div>
                <button type="submit">Save Changes</button>
            </form>
        </div>
    </div>
)}

{/*Editing Quantity*/}
{isEditingQuantity && (
    <div className="edit-quantity-modal">
        <div className="modal-overlay" onClick={() => setIsEditingQuantity(false)}></div>
        <div className="edit-quantity-form">
            <h3>Edit Quantity</h3>
            <form onSubmit={(e) => { e.preventDefault(); handleUpdateQuantity(); }}>
                <div className="form-group">
                    <label htmlFor="quantity">Quantity</label>
                    <input
                        type="number"
                        id="quantity"
                        value={editItemQuantity}
                        onChange={(e) => setEditItemQuantity(e.target.value)}
                    />
                    {itemQuantityError && <div className="error-message">{itemQuantityError}</div>}
                </div>
                <button type="submit">Save Changes</button>
                <button type="button" onClick={() => setIsEditingQuantity(false)}>Cancel</button>
            </form>
        </div>
    </div>
)}
{/* Full-Screen Discount Edit Modal */}
{isEditingDiscount && (
    <div className="edit-discount-modal">
        <div className="modal-overlay" onClick={() => setIsEditingDiscount(false)}></div>
        <div className="edit-discount-form">
            <h3>Edit Discount</h3>
            <form onSubmit={(e) => { e.preventDefault(); handleUpdateDiscount(); }}>
                <div className="form-group">
                    <label htmlFor="discount">Discount</label>
                    <input
                        type="number"
                        id="discount"
                        value={editItemDiscount}
                        onChange={(e) => {
                            const discountValue = e.target.value;
                            setEditItemDiscount(discountValue);
                            if (discountValue === '') {
                                setEditItemDiscountDays(0); // Reset discount days if discount is removed
                            }
                        }}
                    />
                            {itemDiscountError && <div className="error-message">{itemDiscountError}</div>}
                            </div>
                {editItemDiscount > 0 && (
                    <div className="form-group">
                        <label htmlFor="discountDays">Discount Expiration (days)</label>
                        <input
                            type="number"
                            id="discountDays"
                            value={editItemDiscountDays}
                            onChange={(e) => setEditItemDiscountDays(e.target.value)}
                        />
                        {itemDiscountDaysError && <div className="error-message">{itemDiscountDaysError}</div>}
                    </div>
                )}
                <button type="submit">Save Changes</button>
                <button type="button" onClick={() => setIsEditingDiscount(false)}>Cancel</button>
            </form>
        </div>
    </div>
)}



  {/* Pagination Controls */}
  <div className="pagination">
    <button onClick={handlePreviousPage}>Previous</button>
    {pages.map((page) => (
      <button key={page} onClick={() => setCurrentPage(page)}>
        {page}
      </button>
    ))}
    {endPage < totalPages && (
      <button onClick={() => setCurrentPage(endPage + 1)}>...</button>
    )}
    <button onClick={handleNextPage}>Next</button>
  </div>

{/* Admins Section */}
{(userRole === '1' || (userRole === '2' && isCurrentUserAdmin)) && (
  <div className="admins-section card">
    <h2 className="section-title">Admins</h2>
    <div className="add-admin-container">
      <input
        type="email"
        value={newAdminEmail}
        onChange={(e) => setNewAdminEmail(e.target.value)}
        placeholder="Admin email"
        className="admin-input"
      />
      <button onClick={handleAddAdmin} className="admin-button">Add Admin</button>
    </div>
    <ul className="admins-list">
      {admins.map((admin) => (
        <li key={admin.user_id} className="admin-item">
          <span>{admin.email}</span>
          <button onClick={() => handleRemoveAdmin(admin.user_id)}>Remove</button>
        </li>
      ))}
    </ul>
    {error && error.includes('admin') && (
      <div className="error-message">{error}</div>
    )}
  </div>
)}

        </div>
      </div>
      {loading && <div className="spinner"></div>}

      </div>
);
}

export default VendorDetails;