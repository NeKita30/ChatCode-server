import streamlit as st
import api

def get_user_list():
    usr_lst = api.send_request("users")
    if usr_lst["ok"] and usr_lst["users"]:
        return usr_lst["users"]
    else:
        return []


st.markdown("""
# Новый чат
""")

with st.form("new_chat"):
    chat_name = st.text_input("Название:")

    usr_lst = get_user_list()
    options = []
    for usr in usr_lst:
        if usr["id"] != st.session_state.user_id:
            options.append(usr)
    
    members = st.multiselect("Добавить в чат:", options=options, 
                   format_func=lambda usr: usr["name"])
    
    submit = st.form_submit_button('Создать чат')

if submit:
    if not chat_name:
        st.error("Название не должно быть пустым")
    if not members:
        st.error("Добавьте хотя бы одного участника")
    if chat_name and members:
        mids = [m["id"] for m in members]
        mids.append(st.session_state.user_id)
        result = api.send_request("make_chat", chatname=chat_name,
                                  membersid=mids)
        if result["ok"]:
            st.switch_page("login.py")

