import streamlit as st
import api
import chat_page

if "user_name" not in st.session_state:
    st.session_state.user_name = None
    st.session_state.user_id = None

def chat_wrapper(chat):
    def make_chat_page():
        chat_page.create_page(chat["name"],
                              chat["id"])
    func = make_chat_page
    func.__name__ = f"make_chat_{chat["id"]}_page"
    return make_chat_page

def get_chats():
    if st.session_state.user_name: 
        result = api.send_request("chats", 
                                  userid=st.session_state.user_id)
        if result["ok"] and result["chats"]:
            return result["chats"]
    return []

readme = st.Page("login.py", title="Главная", default=True, icon=":material/home:")
makechat = st.Page("makechat.py", title="Новый чат", icon=":material/add:")
pages = [st.Page(chat_wrapper(chat), title=chat["name"], icon=":material/forum:") for chat in get_chats()]
pg_list = [readme, *pages]
if st.session_state.user_name:
    pg_list.insert(1, makechat)

pg = st.navigation(pg_list)
st.set_page_config(page_title="ChatCode",
                   page_icon=":material/assignment:")

pg.run()

