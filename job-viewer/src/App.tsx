import { Badge, Button, Card, Drawer, Flex, Form, Input, Modal, notification, Space, Tag } from 'antd';
import './App.css'
import { ClockCircleOutlined, SyncOutlined } from '@ant-design/icons';
import { useEffect, useState } from 'react';
import axios from 'axios';
const JobStatusView = ({ sts }: { sts: string }) => {
  return (
    <>
      {sts.toLowerCase().includes("queue") && (
        <Tag icon={<ClockCircleOutlined />} color="default">
          QUEUED
        </Tag>
      )}
      {sts.toLowerCase().includes("scheduled") && (
        <Tag icon={<SyncOutlined spin />} color="processing">
          processing on worker 1
        </Tag>
      )}
      {sts.toLowerCase().includes("finish") && (
        <Tag color="success">Completed</Tag>
      )}
      {sts.toLowerCase().includes("error") && (
        <Tag color="red">Error</Tag>
      )}
    </>
  );
}
class WorkerStatus{
    max_workers?: number = 0;
    active_workers?: number
    worker_name?: string
}
class MasterStatus{
  master_online: boolean = false
  workers: WorkerStatus[] = []
}
// type Job struct {
//     ID         primitive.ObjectID `bson:"_id,omitempty"`
//     PythonCode string             `bson:"python_code"`
//     Status     string             `bson:"status"`
//     WorkerName string             `bson:"worker_name,omitempty"`
//     Message string               `bson:"message,omitempty"`
//     CreatedAt  time.Time          `bson:"created_at"`

class JobStatus{
  _id: string = ""
  worker_name: string = ""
  status: string = ""
  message: string = ""
  created_at: string = ""
  python_code: string = ""

}
const BACKEND_URL = "http://localhost:8085"
const App = () => {
  const [drawerVisible, setDrawerVisible] = useState<boolean>(false);
  const [form] = Form.useForm();
  const [masterStatus,setMasterStatus] = useState<MasterStatus>();
  const [jobStatus,setJobStatus] = useState<JobStatus[]>([]);
  const showDrawer = () => {
    setDrawerVisible(true);
  };
  const getMasterStatus = async function(){
    try {
      const { data } = await axios.get(BACKEND_URL + "/status");
      console.log("Master Status",data)
      setMasterStatus(data);
    } catch (error) {
      
      console.log("Error connecting to the server");
    }
    
  }
  const [open, setOpen] = useState<boolean>(false);
  const [output, setOutput] = useState<string>("");
  const showModal = (output:string) => {
    setOutput(output);

    setOpen(true);

    
  };
  const getJobStatus = async function(){
    try {
      const { data } = await axios.get(BACKEND_URL + "/jobs");
      console.log("Job Status",data)
      setJobStatus(data);
    } catch (error) {
     console.log("Error connecting to the server",error);
    }
    
  }
  const poll = function(){
    getMasterStatus()
    getJobStatus()
  }
  useEffect(()=>{
    getMasterStatus()
    getJobStatus()
     const interval = setInterval(poll, 3000)
     return ()=>clearInterval(interval)
  },[])
  const [api, contextHolder] = notification.useNotification();
  const closeDrawer = () => {
    setDrawerVisible(false);
  };
  const handleSubmit = async  (values: { python_code: string }) => {
    console.log("Submitted Python Code:", values.python_code);
    // Add your submission logic here
    const {data} = await axios.post(BACKEND_URL + "/submit",{python_code:values.python_code})
    console.log("Submitted",data)
    api.success({
      message: "Job Queued",
      description: "Your job has been queued successfully",
    });
    closeDrawer();
  };
  
  return (
    <div className="App">
      {contextHolder}
       <Modal
       onOk={() => setOpen(false)}
        title={<p>Standard Output / Error</p>}
       cancelButtonProps={{ style: { display: 'none' } }}
        open={open}
        onCancel={() => setOpen(false)}
      >
        <pre>{output}</pre>
      </Modal>
      <header>
        <div className="title">Distrubuted Code Runner</div>
        <div className="title_sub">
          Developed by Rohit Kumar
          <div>
            <Space>
              {!masterStatus && (
                <Badge color="orange" text="Checking master info" />
              )}
              {masterStatus &&
                masterStatus.master_online &&
                masterStatus.workers && (
                  <>
                    {masterStatus.workers
                      .filter((wo) => wo.max_workers !== 0)
                      .map((worker) => {
                        return (
                          <Badge
                            color="green"
                            text={`${worker.worker_name} is active`}
                          />
                        );
                      })}
                  </>
                )}
              {masterStatus &&
                masterStatus.master_online &&
                masterStatus.workers && (
                  <>
                    {masterStatus.workers
                      .filter((wo) => wo.max_workers === 0)
                      .map((worker) => {
                        return (
                          <Badge
                            color="red"
                            text={`${worker.worker_name} is not active`}
                          />
                        );
                      })}
                  </>
                )}
              {masterStatus && !masterStatus.master_online && (
                <Badge color="red" text="Master offline" />
              )}
              <Button onClick={showDrawer}>Add Job</Button>
            </Space>
          </div>
        </div>
      </header>
      <Drawer
        title="Submit New Job"
        width={400}
        onClose={closeDrawer}
        visible={drawerVisible}
        bodyStyle={{ paddingBottom: 80 }}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item
            name="python_code"
            label="Python Code"
            rules={[
              { required: true, message: "Please enter the Python code" },
            ]}
          >
            <Input.TextArea
              rows={6}
              placeholder="Enter your Python code here"
            />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">
              Enqueue
            </Button>
          </Form.Item>
        </Form>
      </Drawer>
      <div className="container">
        <div className="jobs">
          <div className="job_item">
            {jobStatus.length === 0 && (
            <Card bordered={true}>
              <div className="job_title">
                <b> No Jobs Found</b>
              </div>
            </Card>  
            )}
            {jobStatus.map((job) => {
              return (
                <Card bordered={true}>
                  <Flex justify="space-between">
                    <div className="job_title">
                      <b> JOB ID:</b>
                      {job._id}
                    </div>
                    <div className="job_status">
                      {" "}
                      <JobStatusView sts={job.status} />
                    </div>
                    <div className="job_status">
                      {" "}
                      <b> Queued At</b> {new Date(job.created_at).toLocaleString()}
                    </div>
                    <Button type="primary" onClick={()=>{
                      showModal(job.message)
                    }}>Output</Button>
                  </Flex>
                </Card>
              );
            })}
           
          </div>
        </div>
      </div>
    </div>
  );
}
export default App;