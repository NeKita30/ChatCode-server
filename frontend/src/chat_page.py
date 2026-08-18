import streamlit as st
import api

@st.fragment(run_every="10s")
def print_history(members_id, members_name,
                  usr_id, chat_id):
    chat = api.send_request("chat_history", 
                            userid=usr_id,
                            chatid=chat_id)
    history = []
    if chat["ok"] and chat["history"]:
        for msg in chat["history"]:
            msg_d = dict()
            auth_id = msg["author_id"]
            memb_idex = members_id.index(auth_id)
            msg_d["author"] = members_name[memb_idex]
            msg_d["text"] = msg["text"]

            history.append(msg_d)

    with st.container(key="cont_msg", 
                      height=400):
        for message in history:
            with st.chat_message(message['author']):
                st.markdown(message['text'])


def create_page(chat_name, chat_id):
    usr_name = st.session_state.user_name
    usr_id = st.session_state.user_id
    members_id = []
    members_name = []
    result = api.send_request("chat_members", chatid=chat_id)
    if result["ok"]:
        for m in result["members"]:
            members_id.append(m["id"])
            members_name.append(m["name"])

    st.markdown(f"### {chat_name}")
    with st.popover(f":material/info: :material/settings:"):
        st.write(f"{len(members_id)} участников:")
        st.write(f"{", ".join(members_name)}")

    print_history(members_id, members_name, usr_id, chat_id)
    
    msg_txt = st.chat_input("Введите сообщение")
    if msg_txt:
        res = api.send_request("send_msg", authorid=usr_id,
                         chatid=chat_id, text=msg_txt)
        st.rerun()
