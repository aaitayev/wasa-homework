<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import api from '../services/axios.js';

const groups = ref([]);
const isLoading = ref(false);
const errorMsg = ref('');
const router = useRouter();

// State for new group form
const showNewGroupForm = ref(false);
const newGroupName = ref('');
const newGroupMembers = ref('');
const fileInputMap = ref({}); // Maps groupId to file input element ref

async function fetchGroups() {
  errorMsg.value = '';
  isLoading.value = true;
  groups.value = [];
  try {
    const res = await api.get('/conversations');
    const allConvs = res.data || [];
    
    // Filter out raw direct messages, keep only groups
    const validDetails = allConvs.filter(conv => conv.isGroup === true);
      
    // process into ui model
    groups.value = validDetails.map(conv => {
      let title = conv.name || conv.id;
      const dt = conv.lastMessageAt ? new Date(conv.lastMessageAt) : new Date(0);
      
      return {
        id: conv.id,
        title: title,
        participants: conv.participants || [],
        lastActivity: dt,
        snippet: conv.lastMessageText || 'No messages',
      };
    }); // already sorted by backend
    
  } catch (error) {
    if (error.response && error.response.status === 401) {
      router.push('/login');
    } else {
      errorMsg.value = 'Failed to load groups.';
      console.error(error);
    }
  } finally {
    isLoading.value = false;
  }
}

async function createGroup() {
  if (!newGroupName.value.trim() || !newGroupMembers.value.trim()) {
    alert("Please provide a group name and at least one member.");
    return;
  }
  
  const usernames = newGroupMembers.value.split(',').map(s => s.trim()).filter(s => s);
  
  try {
    isLoading.value = true;
    errorMsg.value = '';
    
    // 1. Create group via POST /messages
    const msgRes = await api.post('/messages', {
      text: "Group " + newGroupName.value + " created",
      isGroup: true
    });
    
    const groupId = msgRes.data.conversationId;
    if (!groupId) throw new Error("No groupId returned");
    
    // 2. setGroupName
    await api.put(`/groups/${groupId}/name`, { name: newGroupName.value });
    
    // 3. addToGroup for each member
    for (const member of usernames) {
      try {
        await api.post(`/groups/${groupId}/members`, { memberId: member });
      } catch (err) {
        console.warn("Failed to add member", member, err);
      }
    }
    
    // Reset form and route
    showNewGroupForm.value = false;
    newGroupName.value = '';
    newGroupMembers.value = '';
    router.push('/conversations/' + groupId);
    
  } catch (error) {
    errorMsg.value = 'Failed to create new group.';
    console.error(error);
    isLoading.value = false; // Note: Loading remains if push succeeds, which is visually OK as we unmount
  }
}

async function renameGroup(groupId, currentName) {
  const newName = prompt("Enter new group name:", currentName);
  if (!newName || newName === currentName) return;
  try {
    await api.put(`/groups/${groupId}/name`, { name: newName });
    await fetchGroups();
  } catch (error) {
    alert("Failed to rename group.");
  }
}

async function addMember(groupId) {
  const member = prompt("Enter username to add to the group:");
  if (!member) return;
  try {
    await api.post(`/groups/${groupId}/members`, { memberId: member });
    await fetchGroups();
  } catch (error) {
    if (error.response && error.response.status === 404) {
      alert("User not found or you're not in the group.");
    } else if (error.response && error.response.status === 403) {
      alert("Permission denied.");
    } else {
      alert("Failed to add member.");
    }
  }
}

async function leaveGroup(groupId) {
  if (!confirm("Are you sure you want to leave this group?")) return;
  try {
    await api.post(`/groups/${groupId}/leave`);
    await fetchGroups(); // Reload to remove visually
  } catch (error) {
    alert("Failed to leave group.");
  }
}

function openConversation(id) {
  router.push('/conversations/' + id);
}

function formatTime(dateObj) {
  if (!dateObj || dateObj.getTime() === 0) return '';
  return dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function triggerPhotoUpload(groupId) {
  const input = fileInputMap.value[groupId];
  if (input) input.click();
}

async function handlePhotoChange(event, groupId) {
  const file = event.target.files[0];
  if (!file) return;

  // Validate file type
  if (file.type !== 'image/jpeg' && file.type !== 'image/png') {
    alert("Please upload a PNG or JPEG file.");
    event.target.value = '';
    return;
  }

  // Validate file size limit (5MB)
  if (file.size > 5 * 1024 * 1024) {
    alert("File exceeds maximum allowed size of 5MB.");
    event.target.value = '';
    return;
  }

  const arrayBuffer = await file.arrayBuffer();
  
  try {
    isLoading.value = true;
    await api.put(`/groups/${groupId}/photo`, arrayBuffer, {
      headers: {
        'Content-Type': file.type
      }
    });
    alert("Group photo updated successfully.");
  } catch (error) {
    if (error.response && error.response.status === 413) {
      alert("File exceeds server limits.");
    } else {
      alert("Failed to update group photo.");
    }
    console.error(error);
  } finally {
    isLoading.value = false;
    event.target.value = '';
  }
}

// Bind refs dynamically
function setFileInputRef(el, groupId) {
  if (el) {
    fileInputMap.value[groupId] = el;
  }
}

onMounted(() => {
  fetchGroups();
});
</script>

<template>
  <div class="groups-container mt-3">
    <div class="d-flex justify-content-between align-items-center mb-3 border-bottom pb-2">
      <h2>My Groups</h2>
      <div>
        <button class="btn btn-outline-secondary btn-sm me-2" @click="fetchGroups">Refresh</button>
        <button class="btn btn-primary btn-sm" @click="showNewGroupForm = !showNewGroupForm">
          {{ showNewGroupForm ? 'Cancel' : 'Create Group' }}
        </button>
      </div>
    </div>
    
    <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
    
    <!-- New Group Form -->
    <div v-if="showNewGroupForm" class="card mb-4 border-primary shadow-sm">
      <div class="card-header bg-primary text-white">Create a New Group</div>
      <div class="card-body">
        <div class="mb-3">
          <label class="form-label">Group Name</label>
          <input type="text" class="form-control" v-model="newGroupName" placeholder="e.g. Test Squad">
        </div>
        <div class="mb-3">
          <label class="form-label">Members (comma-separated usernames)</label>
          <input type="text" class="form-control" v-model="newGroupMembers" placeholder="e.g. Alice, Bob, Charlie">
        </div>
        <button class="btn btn-success" @click="createGroup" :disabled="isLoading">Create</button>
      </div>
    </div>
    
    <LoadingSpinner v-if="isLoading" />
    
    <div v-if="!isLoading && groups.length > 0" class="row row-cols-1 row-cols-md-2 g-3">
      <div class="col" v-for="group in groups" :key="group.id">
        <div class="card h-100 shadow-sm border-0 position-relative">
          
          <div class="card-body pb-0">
            <!-- Header Area -->
            <div class="d-flex w-100 align-items-center justify-content-between mb-2">
              <h5 class="card-title text-primary mb-0 text-truncate" style="cursor: pointer; max-width: 70%;" @click="openConversation(group.id)">
                {{ group.title }}
              </h5>
              <small class="text-nowrap text-muted border px-1 rounded">{{ formatTime(group.lastActivity) }}</small>
            </div>
            
            <p class="card-text text-muted mb-2 small text-truncate" style="opacity: 0.85;">
              <span class="fw-semibold">Participants:</span> {{ group.participants.join(', ') }}
            </p>
            <p class="card-text mb-3 text-truncate bg-light p-2 rounded small fst-italic">
              "{{ group.snippet }}"
            </p>
          </div>
          
          <!-- Actions Footer -->
          <div class="card-footer bg-white border-top-0 pt-0 d-flex flex-wrap gap-1">
            <button class="btn btn-sm btn-outline-primary" @click="openConversation(group.id)">Chat</button>
            <button class="btn btn-sm btn-outline-secondary" @click="renameGroup(group.id, group.title)">Rename</button>
            <button class="btn btn-sm btn-outline-secondary" @click="addMember(group.id)">Add Member</button>
            <button class="btn btn-sm btn-outline-danger" @click="leaveGroup(group.id)">Leave</button>
            
            <!-- Hidden file input for Photo Upload -->
            <input 
               type="file" 
               class="d-none" 
               accept="image/png, image/jpeg" 
               :ref="el => setFileInputRef(el, group.id)" 
               @change="e => handlePhotoChange(e, group.id)"
            />
            <button class="btn btn-sm btn-outline-info ms-auto" @click="triggerPhotoUpload(group.id)">Upload Photo</button>
          </div>
          
        </div>
      </div>
    </div>
    
    <div v-else-if="!isLoading && groups.length === 0" class="text-center text-muted my-5 p-5 bg-light rounded-3">
      <svg class="feather mb-3" style="width: 48px; height: 48px;"><use href="/feather-sprite-v4.29.0.svg#users"/></svg>
      <h5>No groups joined</h5>
      <p>You haven't been added to any groups or haven't created one yet.</p>
      <button class="btn btn-primary mt-2" @click="showNewGroupForm = true">Create your first group</button>
    </div>
    
  </div>
</template>

<style scoped>
</style>
