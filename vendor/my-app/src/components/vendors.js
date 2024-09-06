import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import '../css/vendor.css';
import defaultImage from './vendor.jpg';

function Vendors() {
  const [vendors, setVendors] = useState([]);
  const [filteredVendors, setFilteredVendors] = useState([]);
  const [error, setError] = useState(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);
  const [userVendorId, setUserVendorId] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortOrder, setSortOrder] = useState('latest');
  const [showDropdown, setShowDropdown] = useState(false);
  const [showSearchInput] = useState(false);

  const itemsPerPage = 12;
  const navigate = useNavigate();

  const fetchVendors = useCallback(async (page) => {
    setLoading(true);
    try {
      const response = await fetch(`http://localhost:8080/vendors?page=${page}&pageSize=${itemsPerPage}&sort=${sortOrder}`);
      const data = await response.json();
      if (data && data.Vendors) {
        let vendorsList = data.Vendors;

        // Filter vendors if the user is a vendor
        if (userVendorId && !isAdmin) {
          vendorsList = vendorsList.filter(vendor => vendor.id === userVendorId);
        } else if (userVendorId) {
          // If the user is an admin, prioritize user's vendor
          const userVendorIndex = vendorsList.findIndex(vendor => vendor.id === userVendorId);
          if (userVendorIndex !== -1) {
            const [userVendor] = vendorsList.splice(userVendorIndex, 1);
            vendorsList.unshift(userVendor);
          }
        }

        setVendors(vendorsList);
        setFilteredVendors(vendorsList);
        setTotalPages(Math.ceil(data.TotalCount / itemsPerPage));
      } else {
        console.log('No vendors data found');
      }
    } catch (error) {
      console.error('Error fetching vendors:', error);
      setError('Failed to load vendors');
    }
    setLoading(false);
  }, [sortOrder, userVendorId, isAdmin]);

  const fetchUser = useCallback(async () => {
    const token = localStorage.getItem('token');
    if (!token) {
      return;
    }
  
    try {
      // Fetch user data
      const response = await fetch('http://localhost:8080/me', {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
  
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
  
      const userData = await response.json();
      console.log('User Data:', userData); // Debug user data
  
      if (userData && userData.me) {
        setIsAdmin(userData.me.user_role === "1");
  
        if (userData.me.user_role === "2") {
          const vendorResponse = await fetch(`http://localhost:8080/uservendors/${userData.me.user_info.id}`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });
  
          if (!vendorResponse.ok) {
            throw new Error(`HTTP error! status: ${vendorResponse.status}`);
          }
  
          const vendorData = await vendorResponse.json();
          console.log('Vendor Data:', vendorData); // Debug vendor data
  
          if (vendorData && vendorData.vendor && vendorData.vendor.length > 0) {
            setUserVendorId(vendorData.vendor[0]); // Set the vendor ID
          }
        }
      } else {
        console.log('No user data found');
      }
    } catch (error) {
      console.error('Error fetching user data:', error);
      setError('Failed to load user data');
    }
  }, []);
  
  useEffect(() => {
    fetchVendors(page);
    fetchUser();
  }, [fetchVendors, fetchUser, page]);

  const handlePageChange = (newPage) => {
    if (newPage > 0 && newPage <= totalPages) {
      setPage(newPage);
    }
  };

  const handleEditClick = (vendorId) => {
    if (isAdmin || (userVendorId && vendorId === userVendorId)) {
      navigate(`/edit-vendor/${vendorId}`);
    } else {
      console.log('Non-admin user cannot edit vendors');
    }
  };

  const handleSearchChange = (event) => {
    const searchValue = event.target.value;
    setSearchTerm(searchValue);

    if (searchValue) {
      const results = vendors.filter(vendor => vendor.name.toLowerCase().includes(searchValue.toLowerCase()));
      setFilteredVendors(results);
      setShowDropdown(true);
    } else {
      setFilteredVendors([]); // Hide dropdown when search term is empty
      setShowDropdown(false);
    }
  };

  const handleSortChange = (event) => {
    setSortOrder(event.target.value);
  };

  const handleSelectVendor = (vendorId) => {
    setSearchTerm('');
    setShowDropdown(false);
    navigate(`/vendor/${vendorId}`);
  };

  const handleImageError = (e) => {
    e.target.src = defaultImage;
  };

  const handleAddVendorClick = () => {
    navigate('/add-vendor');
  };

  const handleVisitClick = (vendorId) => {
    navigate(`/vendor/${vendorId}`);
  };

  if (error) {
    return <div className="error-message">{error}</div>;
  }

  return (
    <div className="page-container">
      <div className="vendor-list">
        <h1 className="title">Vendors</h1>

        {/* Search Input Container */}
        <div className={`search-input-container ${showSearchInput ? 'show' : ''}`}>
          <input
            type="text"
            className='search-input'
            placeholder="Search vendors..."
            value={searchTerm}
            onChange={handleSearchChange}
            onFocus={() => searchTerm && setShowDropdown(true)}
            onBlur={() => setTimeout(() => setShowDropdown(false), 200)}
          />
          {showDropdown && searchTerm && filteredVendors.length > 0 && (
            <ul className="dropdown-menu">
              {filteredVendors.map(vendor => (
                <li key={vendor.id} onClick={() => handleSelectVendor(vendor.id)}>
                  {vendor.name}
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* Sort Selection */}
        <div className="sort-selection">
          <select value={sortOrder} onChange={handleSortChange}>
            <option value="latest">Latest</option>
            <option value="name_asc">Name Ascending</option>
            <option value="name_desc">Name Descending</option>
          </select>
        </div>
        {isAdmin && (
          <button className="add-vendor-button" onClick={handleAddVendorClick}>
            Add Vendor
          </button>
        )}
        {/* Vendors List */}
        <ul className="vendor-cards">
          {vendors.length > 0 ? (
            vendors.map(vendor => (
              <li key={vendor.id} className="vendor-card">
                <div className="vendor-header">
                  <img 
                    src={vendor.img || defaultImage}
                    alt={vendor.name}
                    className="vendor-logo"
                    onError={handleImageError}  // Handle image load error
                  />
                  <div className="vendor-content">
                    <h2 className="vendor-name">{vendor.name}</h2>
                    <p className="vendor-description">{vendor.description || 'No description available'}</p>
                    {(isAdmin || (userVendorId && vendor.id === userVendorId)) && (
                      <>
                        <button onClick={() => handleVisitClick(vendor.id)}>Visit</button>
                        <button onClick={() => handleEditClick(vendor.id)}>Edit</button>
                      </>
                    )}
                  </div>
                </div>
              </li>
            ))
          ) : (
            <li>No vendors found</li>
          )}
        </ul>
        <div className="pagination-controls">
          <button onClick={() => handlePageChange(page - 1)} disabled={page === 1 || loading}>Previous</button>
          <span>Page {page} of {totalPages}</span>
          <button onClick={() => handlePageChange(page + 1)} disabled={page === totalPages || loading}>Next</button>
        </div>
        {loading && <p>Loading...</p>}
      </div>
    </div>
  );
}

export default Vendors;
