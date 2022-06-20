import sys
sys.path.insert(0, "..")
import time


from opcua import ua, Server

#With Authentication
users_db = {
    'user1': 'pwd1'
}

def user_manager(session, username, password):
    session.user = Server.user_manager.getter
    return username in users_db and password == users_db[username]

if __name__ == "__main__":

    # setup our server
    server = Server()
    server.set_endpoint("opc.tcp://0.0.0.0:4840/freeopcua/server/")

    #With Certificate！！！
    server.set_security_policy([ua.SecurityPolicyType.Basic256Sha256_SignAndEncrypt])
    # Set Your Private Key And Certificate Path
    server.load_certificate("../cert.der")
    server.load_private_key("../private_key.pem")

    # # With Authentication too!!!
    # server.set_security_IDs(["Username"])
    # server.user_manager.set_user_manager(user_manager)

    # setup our own namespace, not really necessary but should as spec
    uri = "http://examples.freeopcua.github.io"
    idx = server.register_namespace(uri)

    # get Objects node, this is where we should put our nodes
    objects = server.get_objects_node()

    # populating our address space
    myobj = objects.add_object(idx, "MyObject")
    myvar = myobj.add_variable(idx, "MyVariable", 6.7)
    myvar.set_writable()    # Set MyVariable to be writable by clients

    # starting!
    server.start()
    
    try:
        count = 0
        while True:
            time.sleep(1)
            count += 0.1
            myvar.set_value(count)
    finally:
        #close connection, remove subcsriptions, etc
        server.stop()
