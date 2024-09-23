import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import '../css/editvendor.css';
import defaultImage from './vendor.jpg';
import { jwtDecode } from 'jwt-decode';

function EditVendor() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [vendor, setVendor] = useState(null);
  const [image, setImage] = useState(null);
  const [preview, setPreview] = useState(defaultImage);
  const [previousImage, setPreviousImage] = useState(null);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [subscriptionDays, setSubscriptionDays] = useState('');
  const [errorMessages, setErrorMessages] = useState({
    name: '',
    description: '',
    image: '',
    general: '',
    subscriptionDays: '',
  });
  const [fieldErrors, setFieldErrors] = useState({
    name: false,
    description: false,
    image: false,
    subscriptionDays: false,
  });
  const [loading, setLoading] = useState(true);
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

        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

        const data = await response.json();
        if (data && data.vendor) {
          setVendor(data.vendor);
          setName(data.vendor.name);
          setDescription(data.vendor.description);
          setPreview(data.vendor.img || defaultImage);
          setPreviousImage(data.vendor.img || defaultImage);
        } else {
          throw new Error('Vendor data is undefined or null');
        }
      } catch (error) {
        console.error('Error fetching vendor details:', error);
        setErrorMessages((prev) => ({ ...prev, general: 'Failed to load vendor details' }));
      } finally {
        setLoading(false);
      }
    };

    fetchVendorDetails();
  }, [id]);

  const handleFileChange = (event) => {
    const file = event.target.files[0];
    if (file) {
      if (!file.type.match(/^image\/(png|jpg|jpeg|gif)$/)) {
        setErrorMessages((prev) => ({ ...prev, image: 'Please select a valid image file' }));
        setFieldErrors((prev) => ({ ...prev, image: true }));
        setImage(null);
        setPreview(previousImage);
        return;
      } else {
        setErrorMessages((prev) => ({ ...prev, image: '' }));
        setFieldErrors((prev) => ({ ...prev, image: false }));
        setImage(file);

        const reader = new FileReader();
        reader.onloadend = () => {
          setPreview(reader.result);
        };
        reader.readAsDataURL(file);
      }
    } else {
      setErrorMessages((prev) => ({ ...prev, image: '' }));
      setFieldErrors((prev) => ({ ...prev, image: false }));
      setImage(null);
      setPreview(previousImage);
    }
  };

  const handleSave = async (event) => {
    event.preventDefault();
    setFieldErrors({
      name: false,
      description: false,
      image: false,
      subscriptionDays: false,
    });
    setErrorMessages({
      name: '',
      description: '',
      image: '',
      subscriptionDays: '',
      general: '',
    });

    const formData = new FormData();
    formData.append('name', name);
    formData.append('description', description);
    if (subscriptionDays) {
      formData.append('subscriptionDays', subscriptionDays);
    }
    if (image) {
      formData.append('img', image);
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendors/${id}`, {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });

      if (!response.ok) {
        const errorData = await response.json();
        console.error('Backend error details:', errorData);
        if (errorData.error) {
          const errorObject = errorData.error;
          setFieldErrors({
            name: !!errorObject.name,
            description: !!errorObject.description,
            image: !!errorObject.img,
            subscriptionDays: !!errorObject.subscriptionDays,
          });
          setErrorMessages({
            name: errorObject.name || '',
            description: errorObject.description || '',
            image: errorObject.img || '',
            subscriptionDays: errorObject.subscriptionDays || '',
            general: '',
          });
        } else {
          setErrorMessages((prev) => ({ ...prev, general: 'Unexpected error occurred' }));
        }
        return;
      }

      navigate(`/vendor/${id}`);
    } catch (error) {
      console.error('Error updating vendor:', error);
      setErrorMessages((prev) => ({ ...prev, general: 'An unknown error occurred.' }));
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    const confirmDelete = window.confirm('Are you sure you want to delete this vendor?');
    if (!confirmDelete) return;

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/vendors/${id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        console.error('Backend error details:', errorData);
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      navigate('/vendors');
    } catch (error) {
      console.error('Error deleting vendor:', error);
      setErrorMessages((prev) => ({ ...prev, general: 'Failed to delete vendor' }));
    }
  };

  const token = localStorage.getItem('token');
  const userRole = jwtDecode(token).userRole; // Decode the token to get the user role

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="vendor-container">
      <div className="edit-vendor-container">
        <div className="edit-vendor-image-container">
          <input
            type="file"
            ref={imageRef}
            onChange={handleFileChange}
            style={{ display: 'none' }}
          />
          <img
            src={preview}
            alt={vendor?.name}
            className={`vendor-image ${!preview ? 'no-image' : ''}`}
            onClick={() => imageRef.current?.click()}
          />
          {fieldErrors.image && <p className="error-message">{errorMessages.image}</p>}
        </div>
        <div className="edit-vendor-info-container">
          <h1>Edit Vendor</h1>
          {errorMessages.general && <p className="error-message">{errorMessages.general}</p>}
          <form onSubmit={handleSave}>
            <div className="form-group">
              <label htmlFor="name">Name</label>
              <input
                type="text"
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                style={{ borderColor: fieldErrors.name ? 'red' : '' }}
              />
              {errorMessages.name && <p className="error-message">{errorMessages.name}</p>}
            </div>
            <div className="form-group">
              <label htmlFor="description">Description</label>
              <textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                required
                style={{ borderColor: fieldErrors.description ? 'red' : '' }}
              />
              {errorMessages.description && <p className="error-message">{errorMessages.description}</p>}
            </div>
            {userRole === '1' && ( // Render only for admins
              <div className="form-group">
                <label htmlFor="subscriptionDays">Subscription Days (Optional)</label>
                <input
                  type="number"
                  id="subscriptionDays"
                  value={subscriptionDays}
                  onChange={(e) => setSubscriptionDays(e.target.value)}
                  style={{ borderColor: fieldErrors.subscriptionDays ? 'red' : '' }}
                />
                {errorMessages.subscriptionDays && <p className="error-message">{errorMessages.subscriptionDays}</p>}
              </div>
            )}
            <div className="button-group">
              <button type="submit">Save</button>
              {userRole === '1' && (
                <button type="button" className="delete-button" onClick={handleDelete}>Delete</button>
              )}
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export default EditVendor;