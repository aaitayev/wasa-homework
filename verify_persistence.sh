#!/bin/bash

BASE_URL="http://localhost:3000"

if [ "$1" == "verify" ]; then
    ALICE_ID=$2
    BOB_ID=$3
    GROUP_ID=$4
    MSG_ID=$5

    echo "=== PHASE 2: DATA VERIFICATION ==="
    echo "Verifying Alice still has her name..."
    CONVS=$(curl -s -H "Authorization: Bearer $ALICE_ID" $BASE_URL/conversations)
    if echo "$CONVS" | grep -q "$GROUP_ID"; then
        echo "[OK] Group1 persisted."
    else
        echo "[FAIL] Group1 lost. Response: $CONVS"
    fi

    echo "Verifying message and comment..."
    MSG_DATA=$(curl -s -H "Authorization: Bearer $BOB_ID" $BASE_URL/conversations/$GROUP_ID)
    if echo "$MSG_DATA" | grep -q "Hack the planet!"; then
        if echo "$MSG_DATA" | grep -q '"deleted":true'; then
            echo "[OK] Message persisted and remains deleted."
        else
             echo "[FAIL] Message deleted flag lost or not reported correctly. Response: $MSG_DATA"
        fi
        if echo "$MSG_DATA" | grep -q "I agree!"; then
            echo "[OK] Comment persisted."
        else
            echo "[FAIL] Comment lost. Response: $MSG_DATA"
        fi
    else
        echo "[FAIL] Message content lost. Response: $MSG_DATA"
    fi

    echo "Verifying photos..."
    if curl -s -f -H "Authorization: Bearer $ALICE_ID" $BASE_URL/me/photo > /dev/null; then
        echo "[OK] Alice photo persisted."
    else
        echo "[FAIL] Alice photo lost."
    fi

    echo "=== VERIFICATION COMPLETE ==="
else
    echo "=== PHASE 1: DATA CREATION ==="

    # 1. Login Alice & Bob
    echo "Creating users..."
    ALICE_ID=$(curl -s -X POST -d '{"name": "alice"}' -H "Content-Type: application/json" $BASE_URL/session | grep -oE '"identifier":"[^"]+"' | cut -d'"' -f4)
    BOB_ID=$(curl -s -X POST -d '{"name": "bob"}' -H "Content-Type: application/json" $BASE_URL/session | grep -oE '"identifier":"[^"]+"' | cut -d'"' -f4)

    echo "Alice Token: $ALICE_ID"
    echo "Bob Token: $BOB_ID"

    # 2. Create Group
    echo "Creating group..."
    GROUP_RESP=$(curl -s -X POST -H "Authorization: Bearer $ALICE_ID" -H "Content-Type: application/json" -d '{"text": "Hey Bob", "recipient": "bob", "isGroup": true, "name": "Hackers"}' $BASE_URL/messages)
    GROUP_ID=$(echo $GROUP_RESP | grep -oE '"conversationId":"[^"]+"' | cut -d'"' -f4)
    echo "Group ID: $GROUP_ID"

    # 3. Send Message to Group
    echo "Sending message..."
    MSG_RESP=$(curl -s -X POST -H "Authorization: Bearer $ALICE_ID" -H "Content-Type: application/json" -d '{"conversationId": "'$GROUP_ID'", "text": "Hack the planet!"}' $BASE_URL/messages)
    MSG_ID=$(echo $MSG_RESP | grep -oE '"messageId":"[^"]+"' | cut -d'"' -f4)
    echo "Message ID: $MSG_ID"

    # 4. Reaction/Comment
    echo "Commenting..."
    curl -s -X POST -H "Authorization: Bearer $BOB_ID" -H "Content-Type: application/json" -d '{"comment": "I agree!"}' $BASE_URL/messages/$MSG_ID/comment

    # 5. Photos
    echo "Setting photo..."
    echo "ALICE_PHOTO" > /tmp/alice.jpg
    curl -s -X PUT -H "Authorization: Bearer $ALICE_ID" -H "Content-Type: image/jpeg" --data-binary "@/tmp/alice.jpg" $BASE_URL/me/photo

    # 6. Soft Delete
    echo "Deleting message..."
    curl -s -X DELETE -H "Authorization: Bearer $ALICE_ID" $BASE_URL/messages/$MSG_ID

    echo "=== PHASE 1 COMPLETE ==="
    echo "Please RESTART the backend server now."
    echo "Then run this command:"
    echo "./verify_persistence.sh verify \"$ALICE_ID\" \"$BOB_ID\" \"$GROUP_ID\" \"$MSG_ID\""
fi
