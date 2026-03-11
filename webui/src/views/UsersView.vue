<script setup>
import { ref, watch, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import api from '../services/axios.js';
import ErrorMsg from '../components/ErrorMsg.vue';
import LoadingSpinner from '../components/LoadingSpinner.vue';

const router = useRouter();
const searchQuery = ref('');
const users = ref([]);
const isLoading = ref(false);
const errorMsg = ref('');

const myUsername = localStorage.getItem('username') || '';
const photoTimestamp = ref(Date.now());
const isUploading = ref(false);

async function performSearch() {
  if (!searchQuery.value.trim()) {
    users.value = [];
    return;
  }

  errorMsg.value = '';
  isLoading.value = true;
  try {
    const res = await api.get('/users', {
      params: { search: searchQuery.value }
    });
    users.value = res.data || [];
  } catch (error) {
    if (error.response && error.response.status === 401) {
      router.push('/login');
    } else {
      errorMsg.value = 'Failed to search users.';
    }
  } finally {
    isLoading.value = false;
  }
}

async function onMyPhotoUpload(event) {
  const file = event.target.files[0];
  if (!file) return;

  // 1. Client-side validation
  if (file.size > 5 * 1024 * 1024) {
    errorMsg.value = "Photo too large (max 5MB)";
    return;
  }
  if (!['image/jpeg', 'image/png'].includes(file.type)) {
    errorMsg.value = "Invalid file type. Only JPG and PNG allowed.";
    return;
  }

  isUploading.value = true;
  errorMsg.value = '';
  try {
    await api.put('/me/photo', file, {
      headers: { 'Content-Type': file.type }
    });
    // Force refresh of avatars
    photoTimestamp.value = Date.now();
  } catch (error) {
    errorMsg.value = 'Failed to upload profile photo.';
    console.error(error);
  } finally {
    isUploading.value = false;
  }
}

// Debounce logic
let timeoutId = null;
watch(searchQuery, () => {
  if (timeoutId) clearTimeout(timeoutId);
  timeoutId = setTimeout(() => {
    performSearch();
  }, 300);
});

async function doStartChat(targetUsername) {
  errorMsg.value = '';
  isLoading.value = true;
  try {
    // 1. Check existing conversations
    const convsRes = await api.get('/conversations');
    const existing = convsRes.data.find(c => {
      return !c.isGroup && c.participants && c.participants.includes(targetUsername);
    });

    if (existing) {
      router.push(`/conversations/${existing.id}`);
      return;
    }

    // 2. Not found -> Create new DM via POST /messages
    const res = await api.post('/messages', {
      text: "Started a new chat",
      isGroup: false,
      recipient: targetUsername
    });

    if (res.data && res.data.conversationId) {
      router.push(`/conversations/${res.data.conversationId}`);
    } else {
      throw new Error("Missing conversationId in response");
    }
  } catch (error) {
    errorMsg.value = 'Failed to start chat.';
    console.error(error);
  } finally {
    isLoading.value = false;
  }
}

function getPhotoUrl(username) {
  // We use our new endpoint with cache busting
  return `http://localhost:3000/users/${username}/photo?t=${photoTimestamp.value}`;
}

function handleImageError(e) {
  // Fallback if user has no photo (404)
  const name = e.target.alt || 'User';
  e.target.src = `https://ui-avatars.com/api/?name=${name}&background=random`;
}

onMounted(() => {
  // Option: fetch all if search empty? Backend might not support it well.
  // We will leave it empty until user types.
});
</script>

<template>
  <div class="users-container p-3 h-100 bg-light rounded shadow-sm overflow-auto">
    <div class="my-profile-section mb-4 p-4 bg-white border rounded shadow-sm">
      <div class="d-flex align-items-center gap-4">
        <div class="avatar-upload-container position-relative">
          <img 
            :src="getPhotoUrl(myUsername)" 
            @error="handleImageError"
            class="rounded-circle shadow-sm border border-3 border-white" 
            style="width: 100px; height: 100px; object-fit: cover;"
          />
          <label class="upload-btn-overlay position-absolute bottom-0 end-0 bg-primary text-white rounded-circle d-flex align-items-center justify-content-center shadow" style="width: 32px; height: 32px; cursor: pointer;">
            <input type="file" @change="onMyPhotoUpload" hidden accept="image/png, image/jpeg" />
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-camera-fill" viewBox="0 0 16 16">
              <path d="M10.5 8.5a2.5 2.5 0 1 1-5 0 2.5 2.5 0 0 1 5 0z"/>
              <path d="M2 4a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2h-1.172a2 2 0 0 1-1.414-.586l-.828-.828A2 2 0 0 0 9.172 2H6.828a2 2 0 0 0-1.414.586l-.828.828A2 2 0 0 1 3.172 4H2zm.5 2a.5.5 0 1 1 0-1 .5.5 0 0 1 0 1zm9 2.5a3.5 3.5 0 1 1-7 0 3.5 3.5 0 0 1 7 0z"/>
            </svg>
          </label>
        </div>
        <div>
          <h4 class="mb-1 fw-bold">Hello, {{ myUsername }}!</h4>
          <p class="text-muted small mb-0">Customize your profile photo so others can recognize you.</p>
          <div v-if="isUploading" class="mt-2 d-flex align-items-center gap-2 text-primary small">
            <span class="spinner-border spinner-border-sm"></span>
            Uploading...
          </div>
        </div>
      </div>
    </div>

    <div class="explore-section">
      <div class="d-flex justify-content-between align-items-center mb-4 border-bottom pb-3">
        <h2 class="mb-0">Explore Users</h2>
        <button class="btn btn-outline-secondary btn-sm" @click="performSearch" :disabled="isLoading">Refresh</button>
      </div>

      <div class="search-bar mb-4 position-relative">
        <div class="input-group shadow-sm" style="border-radius: 25px; overflow: hidden;">
          <span class="input-group-text bg-white border-end-0 px-3">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" fill="currentColor" class="bi bi-search text-muted" viewBox="0 0 16 16">
              <path d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1 1 0 0 0 1.415-1.414l-3.85-3.85a1.007 1.007 0 0 0-.115-.1zM12 6.5a5.5 5.5 0 1 1-11 0 5.5 5.5 0 0 1 11 0z"/>
            </svg>
          </span>
          <input 
            type="text" 
            class="form-control border-start-0 ps-1 py-2" 
            v-model="searchQuery" 
            placeholder="Search by username..."
            style="box-shadow: none;"
          />
        </div>
      </div>

      <ErrorMsg v-if="errorMsg" :msg="errorMsg" class="mb-3" />
      <LoadingSpinner v-if="isLoading && users.length === 0" :loading="true" class="my-5" />

      <div class="results-list" v-if="users.length > 0">
        <div 
          v-for="user in users" 
          :key="user.name"
          class="user-card d-flex align-items-center justify-content-between p-3 mb-3 bg-white border rounded shadow-sm transition-all"
        >
          <div class="d-flex align-items-center">
            <div class="avatar-container me-3">
              <img 
                :src="getPhotoUrl(user.name)" 
                :alt="user.name" 
                class="rounded-circle shadow-sm" 
                style="width: 50px; height: 50px; object-fit: cover; border: 2px solid #fff;"
                @error="handleImageError"
              />
            </div>
            <div>
              <h5 class="mb-0 fw-bold">{{ user.name }}</h5>
              <small class="text-muted">Available for chat</small>
            </div>
          </div>
          <button class="btn btn-primary px-4 rounded-pill shadow-sm d-flex align-items-center gap-2" @click="doStartChat(user.name)" :disabled="isLoading">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-chat-dots-fill" viewBox="0 0 16 16">
              <path d="M16 8c0 3.866-3.582 7-8 7a9.06 9.06 0 0 1-2.347-.306c-.584.296-1.925.864-4.181 1.234-.2.032-.352-.176-.273-.362.354-.836.674-1.95.77-2.966C.744 11.37 0 9.76 0 8c0-3.866 3.582-7 8-7s8 3.134 8 7zM5 8a1 1 0 1 0-2 0 1 1 0 0 0 2 0zm4 0a1 1 0 1 0-2 0 1 1 0 0 0 2 0zm3 1a1 1 0 1 0 0-2 1 1 0 0 0 0 2z"/>
            </svg>
            Start Chat
          </button>
        </div>
      </div>

      <div v-else-if="!isLoading && searchQuery.length > 0" class="text-center text-muted my-5 p-5 border rounded bg-white dashed">
        <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" fill="currentColor" class="bi bi-person-x text-muted mb-3 opacity-25" viewBox="0 0 16 16">
          <path d="M11 5a3 3 0 1 1-6 0 3 3 0 0 1 6 0ZM8 7a2 2 0 1 0 0-4 2 2 0 0 1 0 4Zm.256 7a4.474 4.474 0 0 1-.229-1.004H3c.001-.246.154-.986.832-1.664C4.484 10.68 5.711 10 8 10c.26 0 .507.009.74.025.226-.341.496-.65.804-.918C9.077 9.038 8.564 9 8 9c-5 0-6 3-6 4s1 1 1 1h5.256Z"/>
          <path d="M12.5 16a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Zm-.646-4.854.646.647.646-.647a.5.5 0 0 1 .708.708l-.647.646.647.646a.5.5 0 0 1-.708.708l-.646-.647-.646.647a.5.5 0 0 1-.708-.708l.647-.646-.647-.646a.5.5 0 0 1 .708-.708Z"/>
        </svg>
        <h5>No users found</h5>
        <p>Try searching for a different username prefix.</p>
      </div>

      <div v-else-if="searchQuery.length === 0 && !isLoading" class="text-center text-muted my-5">
        <p>Type a name above to find other users on WASA.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.users-container {
  max-width: 800px;
  margin: 0 auto;
}

.my-profile-section {
  border-left: 5px solid #0d6efd !important;
}

.avatar-upload-container {
  transition: transform 0.2s;
}

.avatar-upload-container:hover {
  transform: scale(1.05);
}

.upload-btn-overlay {
  transition: all 0.2s;
}

.upload-btn-overlay:hover {
  background-color: #0b5ed7 !important;
  transform: scale(1.1);
}

.user-card {
  transition: all 0.2s ease-in-out;
}

.user-card:hover {
  transform: translateY(-2px);
  border-color: #0d6efd;
}

.dashed {
  border-style: dashed !important;
}

.shadow-inner {
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.05);
}

.avatar-container img {
  border: 2px solid #ddd;
}
</style>
