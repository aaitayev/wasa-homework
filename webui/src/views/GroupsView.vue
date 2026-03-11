<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import api from '../services/axios.js';
import ErrorMsg from '../components/ErrorMsg.vue';
import LoadingSpinner from '../components/LoadingSpinner.vue';

const router = useRouter();
const groups = ref([]);
const allUsers = ref([]);
const selectedUsers = ref([]);
const userSearch = ref('');
const isLoading = ref(false);
const errorMsg = ref('');

// Create Group Modal state
const isCreating = ref(false);
const newGroupName = ref('');

const myUsername = localStorage.getItem('username') || '';

async function loadGroups() {
  errorMsg.value = '';
  isLoading.value = true;
  try {
    const res = await api.get('/conversations');
    // Filter only groups
    groups.value = res.data.filter(c => c.isGroup === true);
  } catch (error) {
    if (error.response && error.response.status === 401) {
      router.push('/login');
    } else {
      errorMsg.value = 'Failed to load groups.';
    }
  } finally {
    isLoading.value = false;
  }
}

async function loadUsers() {
  try {
    const res = await api.get('/users');
    allUsers.value = res.data.filter(u => u.name !== myUsername);
  } catch (error) {
    console.error("Failed to load users", error);
  }
}

const filteredUsers = computed(() => {
  if (!userSearch.value) return allUsers.value;
  return allUsers.value.filter(u => u.name.toLowerCase().includes(userSearch.value.toLowerCase()));
});

function toggleUser(username) {
  const index = selectedUsers.value.indexOf(username);
  if (index === -1) {
    selectedUsers.value.push(username);
  } else {
    selectedUsers.value.splice(index, 1);
  }
}

async function doCreateGroup() {
  if (!newGroupName.value.trim()) return;
  
  errorMsg.value = '';
  isLoading.value = true;
  
  try {
    const res = await api.post('/messages', {
      text: "Group created",
      isGroup: true,
      name: newGroupName.value,
      participants: selectedUsers.value
    });
    
    isCreating.value = false;
    newGroupName.value = '';
    selectedUsers.value = [];
    
    // Navigate to the new group chat
    router.push(`/conversations/${res.data.conversationId}`);
  } catch (error) {
    errorMsg.value = 'Failed to create group.';
  } finally {
    isLoading.value = false;
  }
}

async function doRename(groupId) {
  const newName = prompt("Enter new group name:");
  if (!newName) return;
  
  errorMsg.value = '';
  isLoading.value = true;
  try {
    await api.put(`/groups/${groupId}/name`, { name: newName });
    await loadGroups();
  } catch (error) {
    errorMsg.value = 'Failed to rename group.';
  } finally {
    isLoading.value = false;
  }
}

async function doAddMember(groupId) {
  const username = prompt("Enter username to add:");
  if (!username) return;
  
  errorMsg.value = '';
  isLoading.value = true;
  try {
    await api.post(`/groups/${groupId}/members`, { memberId: username });
    await loadGroups();
  } catch (error) {
    errorMsg.value = 'Failed to add member. Check if username exists.';
  } finally {
    isLoading.value = false;
  }
}

async function doLeave(groupId) {
  if (!confirm("Are you sure you want to leave this group?")) return;
  
  errorMsg.value = '';
  isLoading.value = true;
  try {
    await api.post(`/groups/${groupId}/leave`);
    await loadGroups();
  } catch (error) {
    errorMsg.value = 'Failed to leave group.';
  } finally {
    isLoading.value = false;
  }
}

const groupPhotoTimestamp = ref(Date.now());

async function onPhotoUpload(event, groupId) {
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

  errorMsg.value = '';
  isLoading.value = true;
  try {
    await api.put(`/groups/${groupId}/photo`, file, {
      headers: { 'Content-Type': file.type }
    });
    // Force re-render of images using a timestamp
    groupPhotoTimestamp.value = Date.now();
    await loadGroups();
  } catch (error) {
    errorMsg.value = 'Failed to upload group photo.';
    console.error(error);
  } finally {
    isLoading.value = false;
  }
}

function getGroupPhotoUrl(groupId) {
  return `http://localhost:3000/groups/${groupId}/photo?t=${groupPhotoTimestamp.value}`;
}

function handleImageError(e) {
  e.target.src = 'https://ui-avatars.com/api/?name=Group&background=6c757d&color=fff';
}

function formatTime(dateStr) {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) + ' ' + d.toLocaleDateString();
}

onMounted(() => {
  loadGroups();
  loadUsers();
});
</script>

<template>
  <div class="groups-container p-3 h-100 bg-light rounded shadow-sm overflow-auto">
    <div class="d-flex justify-content-between align-items-center mb-4 border-bottom pb-3">
      <h2 class="mb-0">Group Chats</h2>
      <button class="btn btn-primary d-flex align-items-center gap-2 rounded-pill px-4" @click="isCreating = true">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" fill="currentColor" class="bi bi-plus-lg" viewBox="0 0 16 16">
          <path fill-rule="evenodd" d="M8 2a.5.5 0 0 1 .5.5v5h5a.5.5 0 0 1 0 1h-5v5a.5.5 0 0 1-1 0v-5h-5a.5.5 0 0 1 0-1h5v-5A.5.5 0 0 1 8 2Z"/>
        </svg>
        New Group
      </button>
    </div>

    <ErrorMsg v-if="errorMsg" :msg="errorMsg" class="mb-3" />
    <LoadingSpinner v-if="isLoading && groups.length === 0" :loading="true" />

    <div class="row row-cols-1 row-cols-md-2 g-4" v-if="groups.length > 0">
      <div v-for="group in groups" :key="group.id" class="col">
        <div class="card h-100 border-0 shadow-sm transition-hover">
          <div class="card-body p-4">
            <div class="d-flex align-items-start gap-3 mb-3">
              <div class="avatar-container position-relative">
                <img 
                  :src="getGroupPhotoUrl(group.id)" 
                  @error="handleImageError"
                  class="rounded-3 shadow-sm" 
                  style="width: 64px; height: 64px; object-fit: cover;"
                />
                <label class="photo-upload-overlay position-absolute top-0 start-0 w-100 h-100 d-flex align-items-center justify-content-center rounded-3 bg-dark bg-opacity-50 text-white opacity-0 transition-opacity" style="cursor: pointer;">
                  <input type="file" @change="onPhotoUpload($event, group.id)" hidden accept="image/*" />
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" class="bi bi-camera-fill" viewBox="0 0 16 16">
                    <path d="M10.5 8.5a2.5 2.5 0 1 1-5 0 2.5 2.5 0 0 1 5 0z"/>
                    <path d="M2 4a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2h-1.172a2 2 0 0 1-1.414-.586l-.828-.828A2 2 0 0 0 9.172 2H6.828a2 2 0 0 0-1.414.586l-.828.828A2 2 0 0 1 3.172 4H2zm.5 2a.5.5 0 1 1 0-1 .5.5 0 0 1 0 1zm9 2.5a3.5 3.5 0 1 1-7 0 3.5 3.5 0 0 1 7 0z"/>
                  </svg>
                </label>
              </div>
              <div class="flex-grow-1 overflow-hidden">
                <h5 class="card-title fw-bold mb-1 text-truncate" @click="router.push(`/conversations/${group.id}`)" style="cursor: pointer;">
                  {{ group.name || 'Untitled Group' }}
                </h5>
                <p class="text-muted small mb-1">
                  {{ group.participants ? group.participants.length : 0 }} members: {{ group.participants ? group.participants.join(', ') : '' }}
                </p>
                <div v-if="group.lastMessageText" class="last-msg-preview text-truncate small italic text-secondary">
                  "{{ group.lastMessageText }}" — {{ formatTime(group.lastMessageAt) }}
                </div>
              </div>
            </div>

            <!-- Action Buttons -->
            <div class="d-flex flex-wrap gap-2 border-top pt-3 mt-auto">
              <button class="btn btn-sm btn-outline-primary" @click="doRename(group.id)">Rename</button>
              <button class="btn btn-sm btn-outline-info" @click="doAddMember(group.id)">+ Member</button>
              <button class="btn btn-sm btn-outline-danger" @click="doLeave(group.id)">Leave</button>
              <button class="btn btn-sm btn-primary ms-auto" @click="router.push(`/conversations/${group.id}`)">Chat</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="!isLoading" class="text-center text-muted my-5 py-5 border rounded bg-white dashed">
      <div class="mb-3">
        <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" fill="currentColor" class="bi bi-people text-muted opacity-25" viewBox="0 0 16 16">
          <path d="M15 14s1 0 1-1-1-4-5-4-5 3-5 4 1 1 1 1h8Zm-7.978-1A.261.261 0 0 1 7 12.996c.001-.264.167-1.03.76-1.72C8.312 10.629 9.282 10 11 10c1.717 0 2.687.63 3.24 1.276.593.69.758 1.457.76 1.72l-.008.002a.274.274 0 0 1-.014.002H7.022ZM11 7a2 2 0 1 0 0-4 2 2 0 0 0 0 4Zm3-2a3 3 0 1 1-6 0 3 3 0 0 1 6 0ZM6.936 9.28a5.88 5.88 0 0 0-1.23-.247A7.35 7.35 0 0 0 5 9c-4 0-5 3-5 4 0 .667.333 1 1 1h4.216A2.238 2.238 0 0 1 5 13c0-1.01.377-2.042 1.09-2.904.243-.294.526-.569.846-.816ZM4.92 10A5.493 5.493 0 0 0 4 13H1c0-.26.164-1.03.76-1.724.545-.636 1.487-1.276 3.16-1.276ZM1.5 5.5a3 3 0 1 1 6 0 3 3 0 0 1-6 0Zm3-2a2 2 0 1 0 0 4 2 2 0 0 0 0-4Z"/>
        </svg>
      </div>
      <h5>No groups yet</h5>
      <p>Create your first group to start chatting with multiple friends!</p>
    </div>

    <!-- Create Group Modal (using manual overlay) -->
    <div v-if="isCreating" class="modal-overlay position-fixed top-0 start-0 w-100 h-100 d-flex align-items-center justify-content-center p-3" style="z-index: 2000; background: rgba(0,0,0,0.5);">
      <div class="card shadow w-100" style="max-width: 500px;">
        <div class="card-header bg-white d-flex justify-content-between align-items-center p-3">
          <h5 class="mb-0 fw-bold">Create New Group</h5>
          <button type="button" class="btn-close" @click="isCreating = false"></button>
        </div>
        <div class="card-body p-4">
          <div class="mb-3">
            <label class="form-label fw-bold">Group Name</label>
            <input type="text" class="form-control" v-model="newGroupName" placeholder="e.g., Vacation Friends" />
          </div>
          <div class="mb-4">
            <label class="form-label fw-bold d-flex justify-content-between">
              Select Members
              <span class="badge bg-primary rounded-pill">{{ selectedUsers.length }} selected</span>
            </label>
            <div class="input-group input-group-sm mb-2">
              <span class="input-group-text bg-white border-end-0">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" fill="currentColor" class="bi bi-search" viewBox="0 0 16 16">
                  <path d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1 1 0 0 0 1.415-1.414l-3.85-3.85a1.007 1.007 0 0 0-.115-.1zM12 6.5a5.5 5.5 0 1 1-11 0 5.5 5.5 0 0 1 11 0z"/>
                </svg>
              </span>
              <input type="text" class="form-control border-start-0 ps-0" v-model="userSearch" placeholder="Search users..." />
            </div>
            <div class="user-selection-list border rounded p-2 overflow-auto shadow-sm" style="max-height: 200px; background: #fbfbfb;">
              <div v-for="user in filteredUsers" :key="user.name" 
                   @click="toggleUser(user.name)"
                   class="user-selection-item d-flex align-items-center justify-content-between p-2 mb-1 rounded transition-all"
                   :class="selectedUsers.includes(user.name) ? 'bg-primary bg-opacity-10' : 'hover-bg-light'"
                   style="cursor: pointer;">
                <div class="d-flex align-items-center gap-2">
                  <div class="avatar-sm rounded-circle bg-secondary bg-opacity-25 d-flex align-items-center justify-content-center fw-bold text-secondary" style="width: 28px; height: 28px; font-size: 0.7rem;">
                    {{ user.name.charAt(0).toUpperCase() }}
                  </div>
                  <span class="small fw-semibold">{{ user.name }}</span>
                </div>
                <div class="form-check m-0">
                  <input class="form-check-input" type="checkbox" :checked="selectedUsers.includes(user.name)" @click.stop="toggleUser(user.name)" />
                </div>
              </div>
              <div v-if="filteredUsers.length === 0" class="text-center py-4 text-muted small">
                No users found.
              </div>
            </div>
          </div>
          <div class="d-flex gap-2 justify-content-end">
            <button class="btn btn-light rounded-pill px-4" @click="isCreating = false">Cancel</button>
            <button class="btn btn-primary rounded-pill px-4" @click="doCreateGroup" :disabled="!newGroupName || isLoading">
              <span v-if="isLoading" class="spinner-border spinner-border-sm me-2"></span>
              Create Group
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.transition-hover {
  transition: all 0.2s ease-in-out;
}

.transition-hover:hover {
  transform: translateY(-4px);
  box-shadow: 0 10px 20px rgba(0,0,0,0.1) !important;
}

.avatar-container:hover .photo-upload-overlay {
  opacity: 1;
}

.dashed {
  border-style: dashed !important;
}

.hover-bg-light:hover {
  background-color: #f0f2f5;
}

.user-selection-item {
  transition: all 0.15s ease-in-out;
}

.transition-all {
  transition: all 0.2s ease-in-out;
}
</style>
