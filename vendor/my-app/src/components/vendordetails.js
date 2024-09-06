import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import '../css/vendordetails.css';
import defaultimage from './vendor.jpg';

function VendorDetails() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [vendor, setVendor] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const [userRole, setUserRole] = useState(null);
  const imageRef = useRef(null);

  useEffect(() => {
    const fetchVendorDetails = async () => {
      try {
        const token = localStorage.getItem('token');
        const response = await fetch(`http://localhost:8080/vendors/${id}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          if (response.status === 401) {
            navigate('/login');
            return;
          }
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();

        if (data && data.vendor) {
          setVendor(data.vendor);
        } else {
          throw new Error('Vendor data is undefined or null');
        }
      } catch (error) {
        console.error('Error fetching vendor details:', error);
        setError('Failed to load vendor details');
      } finally {
        setLoading(false);
      }
    };

    const fetchUserRole = async () => {
      try {
        const token = localStorage.getItem('token');
        const response = await fetch('http://localhost:8080/me', {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          if (response.status === 401) {
            navigate('/login');
            return;
          }
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        setUserRole(data.me.user_role);
      } catch (error) {
        console.error('Error fetching user role:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchVendorDetails();
    fetchUserRole();
  }, [id, navigate]);

  const handleEditClick = () => {
    navigate(`/edit-vendor/${id}`);
  };

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
    e.target.src = defaultimage;
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error-message">{error}</div>;
  }

  if (!vendor) {
    return <div>No vendor details available</div>;
  }

  return (
    <div className="page-container">
      <div className="vendor-details-container">
        <div className="vendor-image-container">
          <img
            src={vendor.img || defaultimage}  // Use vendor's image or default image
            alt={vendor.name || 'Vendor'}
            className="vendor-image"
            ref={imageRef}
            onMouseOver={handleImageHover}
            onMouseOut={handleImageLeave}
            onError={handleImageError}  // Handle image load error
          />
        </div>
        <div className="vendor-info-container">
          <h1 className="vendor-name">{vendor.name}</h1>
          <p className="vendor-description">{vendor.description || 'No description available'}</p>
          {userRole === '1' && (
            <button className="edit-button" onClick={handleEditClick}>
              Edit Vendor
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

export default VendorDetails;
