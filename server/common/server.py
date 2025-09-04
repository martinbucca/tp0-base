import socket
import sys
import signal
import logging
import threading 
from common.utils import store_bets, load_bets, has_won
from common.communication import AgencySocket
from common.communication import CHUNK_BET_MESSAGE_ID, FINISH_MESSAGE_ID, GET_WINNERS_MESSAGE_ID


TIMEOUT = 1


class Server:
    def __init__(self, port, listen_backlog, number_of_agencies=5):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._server_socket.settimeout(TIMEOUT)
        self._number_of_agencies = number_of_agencies
        self._agencies_finished = 0
        self._is_currently_running = True
        self._lock = threading.Lock()
        self._winners_by_agency = {}  # agency_id -> [documents of winners]
        self.agencies = []
        self.winners_are_ready = threading.Event()
        
        signal.signal(signal.SIGTERM, self._handle_sigterm)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self._is_currently_running:
            try:
                agency_client_sock = self.__accept_new_connection()
                t = threading.Thread(target=self.__handle_client_connection, args=(agency_client_sock,))
                t.start()
                self.agencies.append(t)

            except OSError:
                if not self._is_currently_running:
                    break 

    def __handle_client_connection(self, agency_client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            # TODO: Modify the receive to avoid short-reads
            while True:
                message_id = agency_client_sock.receive_message_id()
                if message_id == FINISH_MESSAGE_ID:
                    client_id = agency_client_sock.receive_client_id()
                    agency_client_sock.send_finish_message(client_id)
                    logging.info(f"Sending finish message to client {client_id}")
                    with self._lock:
                        self._agencies_finished += 1
                        logging.info(f"action: agencia_finalizo | result: success | total_agencias_finalizadas: {self._agencies_finished}")
                        self.store_winners_for_agency(client_id)
                        if self._agencies_finished == self._number_of_agencies:
                            logging.info("action: sorteo | result: success")
                            self.winners_are_ready.set()
                        break
                elif message_id == CHUNK_BET_MESSAGE_ID:
                    chunk_id, bets_chunk = agency_client_sock.receive_bets_chunk()
                    with self._lock:
                        store_bets(bets_chunk)
                    logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets_chunk)}")
                    agency_client_sock.send_ok_message(chunk_id)
                elif message_id == GET_WINNERS_MESSAGE_ID:
                    logging.info("Solicitud de ganadores")
                    with self._lock:
                        if self._agencies_finished == self._number_of_agencies:
                            logging.info("solicitud de ganadores aceptada. todas las agencias finalizaron")
                            client_id = agency_client_sock.receive_client_id()
                            if self.winners_are_ready.is_set():
                                logging.info("solicitud de ganadores aceptada. todas las agencias finalizaron")
                            winners_list = self._winners_by_agency.get(client_id, [])
                            agency_client_sock.send_winners_list(winners_list)
                        else:
                            logging.info("solicitud de ganadores denegada. faltan agencias por terminar")
                            agency_client_sock.send_no_winners()
                        break
                        

        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            agency_client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return AgencySocket(c)

    def store_winners_for_agency(self, agency_id):
        with self._lock:
            bets = load_bets()
            winners = self._winners_by_agency.get(agency_id, [])
            winners = []
            for bet in bets:
                if bet.agency == agency_id and has_won(bet):
                    winners.append(bet.document)
        self._winners_by_agency[agency_id] = winners
        logging.info(f"action: store_winners | result: success | agency_id: {agency_id} | cantidad: {len(winners)}")

    def shutdown(self):
        try:
            self._is_currently_running = False
            self._server_socket.close()
            logging.info("action: shutdown | result: success | details: server socket closed")
            logging.info("Esperando que terminen todas las conexiones...")
            for t in self.agencies:
                t.join()
            logging.info("Todos los threads finalizaron correctamente.")
        except Exception as e:
            logging.error(f"action: shutdown | result: fail | error: {e}")

    
    def _handle_sigterm(self, *_):
        logging.info("action: shutdown | result: in_progress | reason: SIGTERM received")
        self.shutdown()
        sys.exit(0)

