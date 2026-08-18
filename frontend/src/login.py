import streamlit as st
import api

def registration(new_name, passw):
    if new_name:
        reg_result = api.send_request("registration", 
                                      newname=new_name,
                                      passw=passw)
        if reg_result["ok"]:
            st.session_state.user_name = reg_result["name"]
            st.session_state.user_id = reg_result["id"]
            st.balloons()
            return True
        else:
            st.write(reg_result["msg"])
            return False

def login(user, passw):
    if user:
        log_res = api.send_request("login", 
                                   userid=user["id"], 
                                   passw=passw)
        if log_res["ok"] and log_res["passed"]:
            st.session_state.user_name = user["name"]
            st.session_state.user_id = user["id"]
            return True
        else:
            if not log_res["passed"]:
                st.write("Непавильный пароль")
            else:
                st.write(log_res["msg"])
            return False


def get_user_list():
    usr_lst = api.send_request("users")
    if usr_lst["ok"] and usr_lst["users"]:
        return usr_lst["users"]
    else:
        return []


if st.session_state.user_name is None:
    reg, log = st.tabs(["Регистрация", "Войти"])
    with reg:
        st.header("Регистрация")
        with st.form("reg"):
            new_name = st.text_input("Введите ваше имя:")
            passw = st.text_input("Придумайте пароль:", 
                                  type="password")
            sub = st.form_submit_button("Зарегистрироваться")
        if sub:
            if registration(new_name, passw):
                st.rerun()
    with log:
        st.header("Вход")
        with st.form("log"):
            selected_usr = st.selectbox("Выберите своего пользователя",
                            get_user_list(),
                            format_func=lambda usr: usr["name"],
                            index=None)
            passw = st.text_input("Пароль:", type="password")
            sub = st.form_submit_button("Войти")
        if sub:
            if login(selected_usr, passw):
                st.rerun()
             
else:
    st.write(f"Привет, {st.session_state.user_name}!")
    
    if st.button("Выйти"):
        st.session_state.user_name = None
        st.session_state.user_id = None
        st.rerun()

