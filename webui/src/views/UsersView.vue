<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import api from '../services/axios.js';

const searchQuery = ref('');
const users = ref([]);
const isLoading = ref(false);
const errorMsg = ref('');
const router = useRouter();

// We can implement a simple debounce for live search if we want, 
// for simplicity let's rely on explicit button clicks or simple @input bindings.
let debounceTimer = null;

async function fetchUsers() {
  errorMsg.value = '';
  isLoading.value = true;
  
  try {
    const params = {};
    if (searchQuery.value.trim() !== '') {
      params.search = searchQuery.value.trim();
    }
    const res = await api.get('/users', { params });
    users.value = res.data || [];
  } catch (error) {
    if (error.response && error.response.status === 401) {
      router.push('/login');
    } else {
      errorMsg.value = 'Failed to fetch users.';
      console.error(error);
    }
  } finally {
    isLoading.value = false;
  }
}

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    fetchUsers();
  }, 400); // 400ms debounce
}

async function doStartChat(username) {
  try {
    isLoading.value = true;
    errorMsg.value = '';
    
    // Creating a direct conversation mapping the newly patched backend logic
    const msgRes = await api.post('/messages', {
      text: "Started chat with " + username,
      isGroup: false,
      recipient: username
    });
    
    const conversationId = msgRes.data.conversationId;
    if (conversationId) {
      router.push('/conversations/' + conversationId);
    } else {
      throw new Error("No conversation ID returned by server.");
    }
    
  } catch (error) {
    errorMsg.value = `Failed to start chat with ${username}.`;
    console.error(error);
    isLoading.value = false; // only reset loading on failure as success redirects
  }
}

onMounted(() => {
  fetchUsers();
});
</script>

<template>
  <div class="users-container mt-3">
    <div class="d-flex justify-content-between align-items-center mb-3 border-bottom pb-2">
      <h2>Users</h2>
    </div>
    
    <div class="row mb-4">
      <div class="col-md-8 col-lg-6">
        <label for="searchInput" class="form-label visually-hidden">Search Users</label>
        <div class="input-group">
          <input 
            type="text" 
            id="searchInput"
            class="form-control" 
            placeholder="Search by username..." 
            v-model="searchQuery" 
            @input="onSearchInput"
          >
          <button class="btn btn-outline-primary" type="button" @click="fetchUsers">Search</button>
        </div>
      </div>
    </div>
    
    <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
    <LoadingSpinner v-if="isLoading" />
    
    <div v-if="!isLoading && users.length > 0" class="list-group list-group-flush border rounded-3 p-2 bg-light shadow-sm">
      <div 
        v-for="user in users" 
        :key="user.name"
        class="list-group-item d-flex justify-content-between align-items-center py-3 bg-white mb-1 rounded border-1"
      >
        <span class="fs-5 fw-semibold">{{ user.name }}</span>
        <button 
          class="btn btn-primary btn-sm px-3 rounded-pill" 
          @click="doStartChat(user.name)"
        >
          Start chat
        </button>
      </div>
    </div>
    
    <div v-else-if="!isLoading && users.length === 0" class="text-center text-muted my-5 p-5 bg-light rounded-3">
      <svg class="feather mb-3" style="width: 48px; height: 48px;"><use href="/feather-sprite-v4.29.0.svg#users"/></svg>
      <h5>No users found</h5>
      <p v-if="searchQuery">Try a different search term.</p>
      <p v-else>The system is empty.</p>
    </div>
    
  </div>
</template>

<style scoped>
</style>
