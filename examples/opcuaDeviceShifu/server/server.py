import asyncio
import sys


import time

sys.path.insert(0, "..")
from asyncua import Server
from asyncua import ua
from asyncua.crypto.permission_rules import SimpleRoleRuleset

from asyncua.server.user_managers import CertificateUserManager

#Test Users with UserName Mode,You can add or modify them
users_db = {
    'user1': 'pwd1'
}

def Log(msg):
    #文件地址  __file__，可选添加
    date = time.strftime('%Y.%m.%d %H:%M:%S',time.localtime(time.time()))
    print(date+': '+ msg)

def user_manager(session, username, password):
    session.user = Server.user_manager.getter
    return username in users_db and password == users_db[username]

async def main():
    if len(sys.argv) >= 2 and sys.argv[1] in ["Certificate","UserName","Anonymous"]:
        if sys.argv[1] == "Certificate":
            cert_user_manager = CertificateUserManager()
            await cert_user_manager.add_user("cert.pem", name='test_user')
            server = Server(user_manager=cert_user_manager)
            server.set_security_IDs(["Basic256Sha256"])
        elif sys.argv[1] == "UserName":
            server = Server()
            server.set_security_IDs(["Username"])
        elif sys.argv[1] == "Anonymous":
            server = Server()
            server.set_security_IDs(["Anonymous"])
    else:
        Log("not a valid argument, need to be Certificate/UserName/Anonymous")
        exit()

    server.set_security_policy([ua.SecurityPolicyType.NoSecurity],permission_ruleset=SimpleRoleRuleset)
    await server.init()
    server.set_endpoint("opc.tcp://0.0.0.0:4840/freeopcua/server/")
    Log("server is listening opc.tcp://0.0.0.0:4840/freeopcua/server/ with" + sys.argv[1] + "Mode")
    Log("If you want to cancel the server, please press Ctrl/Control + C")
    server.set_security_policy([ua.SecurityPolicyType.NoSecurity],
                               permission_ruleset=SimpleRoleRuleset())
    # load server certificate and private key. This enables endpoints
    # with signing and encryption.
    idx = 0

    # populating our address space
    myobj = await server.nodes.objects.add_object(idx, "MyObject")
    myvar = await myobj.add_variable(idx, "MyVariable", 0.0)
    await myvar.set_writable()  # Set MyVariable to be writable by clients

    # starting!

    async with server:
        while True:
            await asyncio.sleep(1)
            current_val = await myvar.get_value()
            count = current_val + 0.1
            await myvar.write_value(count)


if __name__ == "__main__":
    asyncio.run(main())
