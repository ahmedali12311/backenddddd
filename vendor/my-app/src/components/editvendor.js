import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import '../css/editvendor.css';
import defaultImage from './vendor.jpg';
import {jwtDecode} from 'jwt-decode';

function EditVendor() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [vendor, setVendor] = useState(null);
  const [image, setImage] = useState(null);
  const [preview, setPreview] = useState(defaultImage);
  const [previousImage, setPreviousImage] = useState(null); // New state for previous image
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [errorMessages, setErrorMessages] = useState({
    name: '',
    description: '',
    image: '',
    general: '',
  });
  const [fieldErrors, setFieldErrors] = useState({
    name: false,
    description: false,
    image: false,
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

        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const data = await response.json();
        console.log('Vendor Data:', data.vendor);

        if (data && data.vendor) {
          setVendor(data.vendor);
          setName(data.vendor.name);
          setDescription(data.vendor.description);
          setPreview(data.vendor.img || defaultImage);
          setPreviousImage(data.vendor.img || defaultImage); // Set previous image
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
        setErrorMessages((prev) => ({
          ...prev,
          image: 'Please select a valid image file',
        }));
        setFieldErrors((prev) => ({ ...prev, image: true }));
        setImage(null);
        setPreview(previousImage); // Reset to previous image
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
      // If no file is selected, reset to previous image
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
    });
    setErrorMessages({
      name: '',
      description: '',
      image: '',
      general: '',
    });

    // Validate image before proceeding
    if (fieldErrors.image) {
      setErrorMessages((prev) => ({ ...prev, general: 'Please correct the image error before saving.' }));
      return;
    }

    const formData = new FormData();
    formData.append('name', name);
    formData.append('description', description);

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
          });
          setErrorMessages({
            name: errorObject.name || '',
            description: errorObject.description || '',
            image: errorObject.img || '',
            general: '',
          });
        } else {
          setErrorMessages((prev) => ({
            ...prev,
            general: 'Unexpected error occurred',
          }));
        }

        return;
      }

      navigate(`/vendor/${id}`);
    } catch (error) {
      console.error('Error updating vendor:', error);
      setErrorMessages((prev) => ({
        ...prev,
        general: 'An unknown error occurred.',
      }));
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
      setErrorMessages((prev) => ({
        ...prev,
        general: 'Failed to delete vendor',
      }));
    }
  };

  const token = localStorage.getItem('token');
  const userRole = jwtDecode(token).userRole;

  if (loading) {
    return <div>Loading...</div>;
  }

  const handleImageError = (e) => {
    if (image) {
      // Only set the error message if the user has uploaded an image
      e.target.src = previousImage || defaultImage;
      setErrorMessages((prev) => ({
        ...prev,
        general: 'Image error: Unable to load image',
      }));
    } else {
      // Reset to previous image if no new image was uploaded
      e.target.src = previousImage || defaultImage;
    }
  };

  // Check if the form should be disabled based on image validity
  const isFormDisabled = fieldErrors.image;

  return (
    <div className="page-container">
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
            onError={handleImageError}
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
            <div className="form-group">
              {errorMessages.image && <p className="error-message">{errorMessages.image}</p>}
            </div>
            <div className="button-group">
              <button type="submit" disabled={isFormDisabled}>Save</button>
              {userRole === 'admin' && (
                <button type="button" onClick={handleDelete}>Delete</button>
              )}
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export default EditVendor;
