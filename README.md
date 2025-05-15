# **XXX Assessment Project**

We are seeking the design of an automatic message sending system for this project.

---

### **Project Description:**

In this project, you are tasked with developing a system that automatically sends 2 messages retrieved from the
database, which have not yet been sent, every 2 minutes, as illustrated in the CURL request provided below.

---

### **Requirements:**

- Retrieve message content, recipient phone number, and sending status for each record from the Database. Character
  limit is required for message content.
- Upon project deployment, automatic message sending should start, processing all unsent records in the database.
- Messages that have been sent once should not be resent. Newly added records should be sent in the subsequent automatic
  process.
- **(Bonus Item)** - After sending a message, cache to Redis the `messageId` value received from the response along with
  the sending time.

---

### **Additional Details:**

- The project should feature 2 separate API endpoints:
    - **Start/Stop** automatic message sending
    - **Retrieve** a list of sent messages
- An example webhook.site has been provided with the Response value for tracking the sent requests.
- **Example webhook.site:** (If the provided link is inactive, you may create your own or seek assistance.)

---

### **Things to Consider:**

- Utilize **Golang** for development.
- **Do not** use any cron package or access the Linux crontab command. We expect you to implement this yourself using
  Go.
- API requests and responses should be documented using [**swagger.io**](https://swagger.io).
- Include a `README.md` file for installation or configuration (e.g., docker commands).
- Host your code on **GitHub** or **Bitbucket**.

---

### **Evaluation Criteria:**

- Architecture / Design patterns / Code Structure
- Code quality / Readability / Consistency / Clean code
- Database/table designs / *(Bonus: Redis usage)*
- The app should work as expected

---

### **Webhook.site Request Payload:**

```bash
curl --location 'https://webhook.site/c3f13233-1ed4-429e-9649-8133b3b9c9cd' --header 'Content-Type: application/json' --header 'x-ins-auth-key: INS.me1x9uMcyYGlhKKQVPoc.bO3j9aZwRTOcA2Ywo' --data '{
  "to": "+905551111111",
  "content": "XXX - project"
}'
```

The webhook.site URL given in the curl example is a sample URL. You can create your own webhook URL via [**webhook.site
**](https://webhook.site). You can set the HTTP response code and body of your webhook as shown in the screenshot.

---

### **Return Value Example:**

**Code:**  
`202`

**Message:**

```json
{
  "message": "Accepted",
  "messageId": "67f2f8a8-ea58-4ed0-a6f9-ff217df4d849"
}
```
