import logging
import sys

import opcua.crypto.security_policies

sys.path.insert(0, "..")
import time


from opcua import ua, Server

#With Authentication
#Test Users, You can add or modify them
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
    if len(sys.argv) >= 2 and sys.argv[1] in ["Certificate","UserName","Anonymous"]:
        if sys.argv[1] == "Certificate":
          print("OPC UA Server is Certificate Mode")
          server.set_security_policy([ua.SecurityPolicyType.Basic256Sha256_SignAndEncrypt])
          server.load_certificate("../cert.der")
          server.load_private_key("../private_key.pem")
        elif sys.argv[1] == "UserName":
          print("OPC UA Server is UserName Mode")
          server.set_security_IDs(["UserName"])
          server.user_manager.set_user_manager(user_manager)
        elif sys.argv[1] == "Anonymous":
          print("OPC UA Server is Anonymous Mode")
    else:
        print("not a valid argument, need to be Certificate/UserName/Anonymous")
        exit(0)

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
