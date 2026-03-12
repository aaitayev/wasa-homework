#!/bin/bash

# Configuration
API_URL=${API_URL:-"http://localhost:3000"}
ALICE="alice"
BOB="bob"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Helper for JSON extraction (works without jq)
extract_json() {
    local key=$1
    local json=$2
    # Simple regex to find "key":"[...] " or "key" : " [...]"
    echo "$json" | grep -o "\"$key\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | head -n1 | cut -d":" -f2- | tr -d ' "'
}

log_pass() { echo -e "${GREEN}[OK]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; exit 1; }

echo "Starting Regression Pass..."

# 1. Login
echo "--- Step 1: Login ---"
res_alice=$(curl -s -X POST "$API_URL/session" -H "Content-Type: application/json" -d "{\"name\": \"$ALICE\"}")
ALICE_TOKEN=$(extract_json "identifier" "$res_alice")
if [ -z "$ALICE_TOKEN" ]; then log_fail "Alice login failed"; fi
log_pass "Alice logged in: $ALICE_TOKEN"

res_bob=$(curl -s -X POST "$API_URL/session" -H "Content-Type: application/json" -d "{\"name\": \"$BOB\"}")
BOB_TOKEN=$(extract_json "identifier" "$res_bob")
if [ -z "$BOB_TOKEN" ]; then log_fail "Bob login failed"; fi
log_pass "Bob logged in: $BOB_TOKEN"

# 2. Messaging
echo "--- Step 2: Messaging ---"
# Alice sends message to Bob (creates DM)
# To create a DM, Alice must search for bob first or just send a message with participants
res_msg=$(curl -s -X POST "$API_URL/messages" \
    -H "Authorization: Bearer $ALICE_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"text\": \"Hello Bob!\", \"participants\": [\"$BOB\"]}")
CONV_ID=$(extract_json "conversationId" "$res_msg")
MSG_ID=$(extract_json "messageId" "$res_msg")

if [ -z "$CONV_ID" ]; then log_fail "Alice failed to send message/create DM"; fi
log_pass "DM created: $CONV_ID, Message ID: $MSG_ID"

# Bob gets conversation
res_get_conv=$(curl -s -X GET "$API_URL/conversations/$CONV_ID" -H "Authorization: Bearer $BOB_TOKEN")
if [[ "$res_get_conv" != *"Hello Bob!"* ]]; then log_fail "Bob cannot see Alice's message"; fi
log_pass "Bob received message successfully"

# Bob replies
res_reply=$(curl -s -X POST "$API_URL/messages" \
    -H "Authorization: Bearer $BOB_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"conversationId\": \"$CONV_ID\", \"text\": \"Hi Alice!\"}")
BOB_MSG_ID=$(extract_json "messageId" "$res_reply")
if [ -z "$BOB_MSG_ID" ]; then log_fail "Bob failed to reply"; fi
log_pass "Bob replied: $BOB_MSG_ID"

# Alice comments
curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/messages/$BOB_MSG_ID/comment" \
    -H "Authorization: Bearer $ALICE_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"comment\": \"Nice reply!\"}" | grep -q "204" || log_fail "Alice failed to comment"
log_pass "Alice commented on Bob's message"

# 3. Soft-delete & 409 Checks
echo "--- Step 3: Deletion & Conflict Checks ---"
# Bob deletes his message
curl -s -o /dev/null -w "%{http_code}" -X DELETE "$API_URL/messages/$BOB_MSG_ID" \
    -H "Authorization: Bearer $BOB_TOKEN" | grep -q "204" || log_fail "Bob failed to delete message"
log_pass "Bob deleted his message"

# Alice tries to comment on deleted message (Should be 409)
http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/messages/$BOB_MSG_ID/comment" \
    -H "Authorization: Bearer $ALICE_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"comment\": \"Wait!\"}")
if [ "$http_code" != "409" ]; then log_fail "Expected 409 for comment on deleted message, got $http_code"; fi
log_pass "Correctly rejected comment on deleted message (409)"

# Alice tries to forward Bob's deleted message (Should be 409)
http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/messages/$BOB_MSG_ID/forward" \
    -H "Authorization: Bearer $ALICE_TOKEN")
if [ "$http_code" != "409" ]; then log_fail "Expected 409 for forward on deleted message, got $http_code"; fi
log_pass "Correctly rejected forward on deleted message (409)"

# 4. Groups
echo "--- Step 4: Groups ---"
res_group=$(curl -s -X POST "$API_URL/messages" \
    -H "Authorization: Bearer $ALICE_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"text\": \"Welcome to the group!\", \"isGroup\": true, \"name\": \"Project X\", \"participants\": [\"$BOB\"]}")
GROUP_ID=$(extract_json "conversationId" "$res_group")
if [ -z "$GROUP_ID" ]; then log_fail "Alice failed to create group"; fi
log_pass "Group created: $GROUP_ID"

# Rename group
curl -s -o /dev/null -w "%{http_code}" -X PUT "$API_URL/groups/$GROUP_ID/name" \
    -H "Authorization: Bearer $ALICE_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"Project Y\"}" | grep -q "204" || log_fail "Alice failed to rename group"
log_pass "Group renamed to Project Y"

# 5. Photos
echo "--- Step 5: Photos ---"
# Create a dummy image
echo "dummy" > dummy.png
res_photo=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$API_URL/me/photo" \
    -H "Authorization: Bearer $ALICE_TOKEN" \
    -H "Content-Type: image/png" \
    --data-binary "@dummy.png")
if [ "$res_photo" != "204" ]; then log_fail "Alice failed to upload photo: $res_photo"; fi
log_pass "Alice uploaded profile photo"

# Get photo and check content type
res_get_photo=$(curl -s -i "$API_URL/me/photo" -H "Authorization: Bearer $ALICE_TOKEN" | grep -i "Content-Type")
if [[ "$res_get_photo" != *"image/png"* ]]; then log_fail "Invalid content-type for photo: $res_get_photo"; fi
log_pass "Photo GET returns correct Content-Type: image/png"

# Group photo
res_group_photo=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$API_URL/groups/$GROUP_ID/photo" \
    -H "Authorization: Bearer $ALICE_TOKEN" \
    -H "Content-Type: image/jpeg" \
    --data-binary "@dummy.png")
if [ "$res_group_photo" != "204" ]; then log_fail "Alice failed to upload group photo: $res_group_photo"; fi
log_pass "Alice uploaded group photo"

rm dummy.png

echo -e "${GREEN}All regression tests passed (except persistence which requires manual restart/validation).${NC}"
exit 0
